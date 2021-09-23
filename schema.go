package seed

import (
	"context"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/generators"
	"golang.org/x/sync/errgroup"
	"sync"

	"github.com/simon-engledew/seed/escape"
)

type Schema map[string][]*Column

func (s Schema) Generator(ctx context.Context, consumer consumers.Consumer) Generator {
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

	return &insertStack{
		ctx:       ctx,
		producers: &sync.WaitGroup{},
		consumers: consumers,
		callback:  callback,
		channels:  channels,
		schema:    s,
		stack:     make(Rows),
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

func (s Schema) Reference(t string, c string) generators.ValueGenerator {
	columns := s[t]
	for idx, column := range columns {
		if column.Name == c {
			return Reference(t, idx)
		}
	}

	return nil
}
