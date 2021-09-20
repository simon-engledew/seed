package seed

import (
	"context"
	"github.com/simon-engledew/seed/distribution"
	"io"
)

type contextKey string

type Generator interface {
	Insert(table string, dist distribution.Distribution, next ...func(Generator))
}

type insertStack struct {
	ctx      context.Context
	cb       func(table string, columns []string, rows chan []string)
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

//func insert(w io.Writer, table TableName, row Row) (int, error) {
//	columns := make([]string, 0, len(row))
//	values := make([]string, 0, len(row))
//
//	for column, value := range row {
//		columns = append(columns, string(column))
//		values = append(values, value)
//	}
//
//	return fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES (%s);\n", table, strings.Join(columns, ","), strings.Join(values, ","))
//}

//func columns(t map[ColumnName]generators.ValueGenerator) []string {
//	out := make([]string, 0, len(t))
//	for c := range t {
//		out = append(out, string(c))
//	}
//	return out
//}

func (i *insertStack) Insert(table string, dist distribution.Distribution, next ...func(Generator)) {
	columns := i.schema[table]

	channel := i.channels[table]

	ctx := context.WithValue(i.ctx, parentKey, i.stack)

	for dist() {
		row := make(Row, 0, len(columns))

		for _, column := range columns {
			row = append(row, column.Generator.Value(ctx))
		}

		channel <- row

		stack := merge(i.stack, Rows{table: row})

		for _, fn := range next {
			fn(&insertStack{
				ctx:      i.ctx,
				channels: i.channels,
				w:        i.w,
				schema:   i.schema,
				stack:    stack,
			})
		}
	}
}
