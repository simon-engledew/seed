package seed

import (
	"context"
	"sync"

	"github.com/simon-engledew/seed/escape"
)

type Schema map[string][]*Column

func (s Schema) Generator(ctx context.Context, cb func(table string, columns []string, rows chan []string)) Generator {
	channels := make(map[string]chan []string)
	for t, columns := range s {
		channel := make(chan []string)
		channels[t] = channel

		names := make([]string, 0, len(columns))

		for _, column := range columns {
			names = append(names, escape.QuoteIdentifier(column.Name))
		}

		cb(t, names, channel)
	}
	return &insertStack{
		wg:       &sync.WaitGroup{},
		ctx:      ctx,
		cb:       cb,
		channels: channels,
		schema:   s,
		stack:    make(Rows),
	}
}

type SchemaTransform func(t string, c *Column)

func (s Schema) Transform(transforms ...SchemaTransform) {
	for t, columns := range s {
		for _, c := range columns {
			for _, transform := range transforms {
				transform(t, c)
			}
		}
	}
}

func (s Schema) Merge(other Schema) {
	for t, columns := range other {
		for c, g := range columns {
			s[t][c] = g
		}
	}
}
