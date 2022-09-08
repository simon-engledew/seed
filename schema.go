package seed

import (
	"context"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/generators"
	"golang.org/x/sync/errgroup"
)

type Column struct {
	Name      string
	Type      string
	Generator generators.ValueGenerator
}

// Schema maps tables to columns.
type Schema map[string][]*Column

// Generator will emit test data to consumer.
func (s Schema) Generator(ctx context.Context, consumer consumers.Consumer) *RowGenerator {
	producers, _ := errgroup.WithContext(ctx)
	consumers, consumersCtx := errgroup.WithContext(ctx)

	callback := consumer(consumersCtx, consumers)

	channels := make(map[string]chan []*generators.Value)
	for t, columns := range s {
		channel := make(chan []*generators.Value)
		channels[t] = channel

		names := make([]string, 0, len(columns))

		for _, column := range columns {
			names = append(names, column.Name)
		}

		callback(t, names, channel)
	}

	return &RowGenerator{
		ctx:       ctx,
		producers: producers,
		consumers: consumers,
		callback:  callback,
		channels:  channels,
		schema:    s,
		stack:     make(Rows),
	}
}

type SchemaTransform func(t string, c *Column)

// ReplaceColumnType will replace the generator for columns with a Type matching t.
func ReplaceColumnType(t string, g generators.ValueGenerator) SchemaTransform {
	return func(table string, column *Column) {
		if column.Type == t {
			column.Generator = g
		}
	}
}

// ReplaceColumns will replace columns found in any table if the name is found in the columns map.
func ReplaceColumns(columns map[string]generators.ValueGenerator) SchemaTransform {
	return func(table string, column *Column) {
		if g, ok := columns[column.Name]; ok {
			column.Generator = g
		}
	}
}

// Merge will replace columns if their table and column names are found in the schema map.
func Merge(schema map[string]map[string]generators.ValueGenerator) SchemaTransform {
	return func(table string, column *Column) {
		if columns, ok := schema[table]; ok {
			if g, ok := columns[column.Name]; ok {
				column.Generator = g
			}
		}
	}
}

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

// Reference produces a ValueGenerator which will emit the value of the parent row.
func (s Schema) Reference(t string, c string) generators.ValueGenerator {
	columns := s[t]
	for idx, column := range columns {
		if column.Name == c {
			return Reference(t, idx)
		}
	}

	return nil
}
