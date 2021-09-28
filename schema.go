package seed

import (
	"context"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/generators"
	"github.com/simon-engledew/seed/types"
	"golang.org/x/sync/errgroup"
	"sync"

	"github.com/simon-engledew/seed/escape"
)

// Schema maps tables to columns.
type Schema map[string][]*Column

// Generator will emit test data to consumer.
func (s Schema) Generator(ctx context.Context, consumer consumers.Consumer) *RowGenerator {
	consumers, ctx := errgroup.WithContext(ctx)

	callback := consumer(consumers)

	channels := make(map[string]chan []string)
	for t, columns := range s {
		channel := make(chan []string)
		channels[t] = channel

		names := make([]string, 0, len(columns))

		for _, column := range columns {
			names = append(names, escape.QuoteIdentifier(column.Name))
		}

		callback(t, names, channel)
	}

	return &RowGenerator{
		ctx:       ctx,
		producers: &sync.WaitGroup{},
		consumers: consumers,
		callback:  callback,
		channels:  channels,
		schema:    s,
		stack:     make(types.Rows),
	}
}

type SchemaTransform func(t string, c *Column)

// Transform iterates through the tables and columns in this schema, calling transforms on each.
func (s Schema) Transform(transforms ...SchemaTransform) {
	for t, columns := range s {
		for _, c := range columns {
			for _, transform := range transforms {
				transform(t, c)
			}
		}
	}
}

// Reference produces a ColumnGenerator which will emit the value of the parent row.
func (s Schema) Reference(t string, c string) generators.ColumnGenerator {
	columns := s[t]
	for idx, column := range columns {
		if column.Name == c {
			return generators.Reference(t, idx)
		}
	}

	return nil
}
