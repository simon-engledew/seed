package seed

import (
	"context"
	"github.com/simon-engledew/seed/consumers"
	"sync"

	"github.com/simon-engledew/seed/escape"
)

type Schema map[string][]*Column

func (s Schema) Generator(ctx context.Context, consumer consumers.Consumer) Generator {
	consumers := &sync.WaitGroup{}

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
