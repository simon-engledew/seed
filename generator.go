package seed

import (
	"context"
	"github.com/simon-engledew/seed/distribution"
	"golang.org/x/sync/errgroup"
	"io"
	"sync"
)

type contextKey string

type Generator struct {
	producers *sync.WaitGroup
	consumers *errgroup.Group
	ctx       context.Context
	callback  func(table string, columns []string, rows chan []string)
	channels  map[string]chan []string
	w         io.Writer
	schema    Schema
	stack     Rows
}

func merge(a Rows, b Rows) Rows {
	copied := make(Rows, len(a)+len(b))
	for k, v := range a {
		copied[k] = v
	}
	for k, v := range b {
		copied[k] = v
	}
	return copied
}

func (g *Generator) Wait() error {
	g.producers.Wait()
	for _, channel := range g.channels {
		close(channel)
	}
	return g.consumers.Wait()
}

func (g *Generator) Insert(table string, dist distribution.Distribution, next ...func(*Generator)) {
	g.InsertContext(g.ctx, table, dist, next...)
}

func (g *Generator) InsertContext(ctx context.Context, table string, dist distribution.Distribution, next ...func(*Generator)) {
	g.producers.Add(1)
	go func() {
		defer g.producers.Done()

		columns := g.schema[table]

		channel := g.channels[table]

		withStack := context.WithValue(ctx, parentKey, g.stack)

		for dist() {
			row := make(Row, 0, len(columns))

			for _, column := range columns {
				row = append(row, column.Generator.Value(withStack))
			}

			channel <- row

			stack := merge(g.stack, Rows{table: row})

			for _, fn := range next {
				fn(&Generator{
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
