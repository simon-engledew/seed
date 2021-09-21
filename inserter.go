package seed

import (
	"context"
	"github.com/simon-engledew/seed/distribution"
	"io"
	"sync"
)

type contextKey string

type Generator interface {
	InsertContext(ctx context.Context, table string, dist distribution.Distribution, next ...func(Generator))
	Insert(table string, dist distribution.Distribution, next ...func(Generator))
	Done()
}

type insertStack struct {
	wg       *sync.WaitGroup
	ctx      context.Context
	consumer func(table string, columns []string, rows chan []string)
	channels map[string]chan []string
	w        io.Writer
	schema   Schema
	stack    Rows
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

func (i *insertStack) Done() {
	i.wg.Wait()
	for _, channel := range i.channels {
		close(channel)
	}
}

func (i *insertStack) Insert(table string, dist distribution.Distribution, next ...func(Generator)) {
	i.InsertContext(i.ctx, table, dist, next...)
}

func (i *insertStack) InsertContext(ctx context.Context, table string, dist distribution.Distribution, next ...func(Generator)) {
	i.wg.Add(1)
	go func() {
		defer i.wg.Done()

		columns := i.schema[table]

		channel := i.channels[table]

		withStack := context.WithValue(ctx, parentKey, i.stack)

		for dist() {
			row := make(Row, 0, len(columns))

			for _, column := range columns {
				row = append(row, column.Generator.Value(withStack))
			}

			channel <- row

			stack := merge(i.stack, Rows{table: row})

			for _, fn := range next {
				fn(&insertStack{
					wg:       i.wg,
					ctx:      ctx,
					channels: i.channels,
					w:        i.w,
					schema:   i.schema,
					stack:    stack,
				})
			}
		}
	}()
}
