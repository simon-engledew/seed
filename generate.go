package seed

import (
	"context"
	"fmt"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/distribution"
	"github.com/simon-engledew/seed/generators"
	"golang.org/x/sync/errgroup"
	"io"
)

type RowGenerator struct {
	producers *errgroup.Group
	consumers *errgroup.Group
	ctx       context.Context
	callback  func(t string, c []string, rows chan []consumers.Value)
	channels  map[string]chan []consumers.Value
	w         io.Writer
	schema    Schema
	stack     consumers.Rows
}

func merge(a consumers.Rows, b consumers.Rows) consumers.Rows {
	copied := make(consumers.Rows, len(a)+len(b))
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

func (g *RowGenerator) Generate(ctx context.Context, columns []*Column) consumers.Row {
	row := make(consumers.Row, 0, len(columns))

	withRow := generators.WithSiblings(ctx, &row)

	for _, column := range columns {
		row = append(row, column.Generator.Value(withRow))
	}

	return row
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

		columnNames := make([]string, 0, len(columns))
		for _, column := range columns {
			columnNames = append(columnNames, column.Name)
		}

		generateCtx := generators.WithColumns(generators.WithParents(ctx, g.stack), columnNames)

		for dist() {
			row := g.Generate(generateCtx, columns)

			channel <- row

			stack := merge(g.stack, consumers.Rows{table: row})

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
