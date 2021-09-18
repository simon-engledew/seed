package seed

import (
	"context"
	"fmt"
	"github.com/simon-engledew/seed/distribution"
	"io"
	"strings"
)

type contextKey string

type Generator interface {
	Insert(table TableName, dist distribution.Distribution, next ...func(Generator))
}

type insertStack struct {
	w      io.Writer
	schema Schema
	stack  Rows
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

func insert(w io.Writer, table TableName, row Row) (int, error) {
	columns := make([]string, 0, len(row))
	values := make([]string, 0, len(row))

	for column, value := range row {
		columns = append(columns, string(column))
		values = append(values, value)
	}

	return fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES (%s);\n", table, strings.Join(columns, ","), strings.Join(values, ","))
}

func (i *insertStack) Insert(table TableName, dist distribution.Distribution, next ...func(Generator)) {
	generators := i.schema[table]

	for dist() {
		row := make(Row, len(generators))

		for column, generator := range generators {
			ctx := context.Background()

			if d, ok := generator.(References); ok {
				ctx = d.WithParents(ctx, i.stack)
			}

			row[column] = generator.Value(ctx)
		}

		_, err := insert(i.w, table, row)
		if err != nil {
			panic(err)
		}

		for _, fn := range next {
			fn(&insertStack{
				w:      i.w,
				schema: i.schema,
				stack:  merge(i.stack, Rows{table: row}),
			})
		}
	}
}
