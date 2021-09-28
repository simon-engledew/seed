package seed

import (
	"context"
	"github.com/simon-engledew/seed/distribution"
	"github.com/simon-engledew/seed/generators"
	"github.com/simon-engledew/seed/types"
	"golang.org/x/sync/errgroup"
	"io"
	"sync"
)

type RowGenerator struct {
	producers *sync.WaitGroup
	consumers *errgroup.Group
	ctx       context.Context
	callback  func(table string, columns []string, rows chan []string)
	channels  map[string]chan []string
	w         io.Writer
	schema    Schema
	stack     types.Rows
}

func merge(a types.Rows, b types.Rows) types.Rows {
	copied := make(types.Rows, len(a)+len(b))
	for k, v := range a {
		copied[k] = v
	}
	for k, v := range b {
		copied[k] = v
	}
	return copied
}

func (g *RowGenerator) Wait() error {
	g.producers.Wait()
	for _, channel := range g.channels {
		close(channel)
	}
	return g.consumers.Wait()
}

func (g *RowGenerator) Insert(table string, dist distribution.Distribution, next ...func(*RowGenerator)) {
	g.InsertContext(g.ctx, table, dist, next...)
}

func (g *RowGenerator) InsertContext(ctx context.Context, table string, dist distribution.Distribution, next ...func(*RowGenerator)) {
	g.producers.Add(1)
	go func() {
		defer g.producers.Done()

		columns := g.schema[table]

		channel := g.channels[table]

		withStack := generators.WithParents(ctx, g.stack)

		for dist() {
			row := make(types.Row, 0, len(columns))

			for _, column := range columns {
				row = append(row, column.Generator.Value(withStack))
			}

			channel <- row

			stack := merge(g.stack, types.Rows{table: row})

			for _, fn := range next {
				fn(&RowGenerator{
					producers: g.producers,
					ctx:       ctx,
					channels:  g.channels,
					w:         g.w,
					schema:    g.schema,
					stack:     stack,
				})
			}
		}
	}()
}
