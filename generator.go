package seed

import (
	"context"
	"fmt"
	"github.com/simon-engledew/seed/distribution"
	"golang.org/x/sync/errgroup"
	"io"
)

type Row []string
type Rows map[string]Row

type RowGenerator struct {
	producers *errgroup.Group
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

func mergeErr(a, b error) error {
	if a != nil {
		if b != nil {
			return fmt.Errorf("%s: %w", a, b)
		}
		return a
	}
	return b
}

func (g *RowGenerator) Wait() error {
	err := g.producers.Wait()
	for _, channel := range g.channels {
		close(channel)
	}
	return mergeErr(g.consumers.Wait(), err)
}

func (g *RowGenerator) Insert(table string, dist distribution.Distribution, next ...func(*RowGenerator)) {
	g.InsertContext(g.ctx, table, dist, next...)
}

func (g *RowGenerator) InsertContext(ctx context.Context, table string, dist distribution.Distribution, next ...func(*RowGenerator)) {
	g.producers.Go(func() error {
		columns, ok := g.schema[table]
		if !ok {
			return fmt.Errorf("unknown table %s", table)
		}

		channel := g.channels[table]

		withStack := WithParents(ctx, g.stack)

		for dist() {
			row := make(Row, 0, len(columns))

			for _, column := range columns {
				row = append(row, column.Generator.Value(withStack))
			}

			channel <- row

			stack := merge(g.stack, Rows{table: row})

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

		return nil
	})
}
