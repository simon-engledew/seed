package seed

import (
	"context"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed/distribution"
	"io"
	"math/rand"
	"strconv"
	"strings"
)

type ColumnGenerator interface {
	Value(ctx context.Context) string
}

type Dependent interface {
	DependsOn() TableName
}

type TableName string
type ColumnName string
type Schema map[TableName]map[ColumnName]ColumnGenerator

type staticColumn struct {
	value string
}

func (c *staticColumn) Value(ctx context.Context) string {
	return c.value
}

func newStaticColumn(value string) ColumnGenerator {
	return &staticColumn{value: value}
}

type callableColumn struct {
	value func() string
}

func (c *callableColumn) Value(ctx context.Context) string {
	return c.value()
}

func newCallableColumn(value func() string) ColumnGenerator {
	return &callableColumn{value: value}
}

type primaryColumn struct {
	count uint64
}

func (c *primaryColumn) Value(ctx context.Context) string {
	c.count += 1
	return strconv.FormatUint(c.count, 10)
}

func PrimaryKey() ColumnGenerator {
	return &primaryColumn{}
}

type contextKey string

var parentKey contextKey = "parent"

type mappedValues map[ColumnName]string

type dependentColumn struct {
	tableName  TableName
	columnName ColumnName
}

func (c *dependentColumn) DependsOn() TableName {
	return c.tableName
}

func (c *dependentColumn) Value(ctx context.Context) string {
	parent := ctx.Value(parentKey).(mappedValues)
	return parent[c.columnName]
}

func Reference(tableName TableName, columnName ColumnName) ColumnGenerator {
	return &dependentColumn{
		tableName:  tableName,
		columnName: columnName,
	}
}

type Insert interface {
	Insert(table TableName, dist distribution.Distribution, next ...func(Insert))
}

type insertStack struct {
	w      io.Writer
	schema Schema
	stack  map[TableName]mappedValues
}

func merge(a map[TableName]mappedValues, b map[TableName]mappedValues) map[TableName]mappedValues {
	copied := make(map[TableName]mappedValues, len(a)+len(b))
	for k, v := range a {
		copied[k] = v
	}
	for k, v := range b {
		copied[k] = v
	}
	return copied
}

func (i *insertStack) Insert(table TableName, dist distribution.Distribution, next ...func(Insert)) {
	generators := i.schema[table]

	for dist() {
		row := make(mappedValues, len(generators))

		for column, generator := range generators {
			ctx := context.Background()

			if d, ok := generator.(Dependent); ok {
				parent := d.DependsOn()

				ctx = context.WithValue(ctx, parentKey, i.stack[parent])
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
				stack:  merge(i.stack, map[TableName]mappedValues{table: row}),
			})
		}
	}
}

func Generate(w io.Writer, schema Schema) Insert {
	gofakeit.Seed(0)
	rand.Seed(0)

	return &insertStack{
		w:      w,
		schema: schema,
		stack:  make(map[TableName]mappedValues, 0),
	}
}

func insert(w io.Writer, table TableName, row map[ColumnName]string) (int, error) {
	columns := make([]string, 0, len(row))
	values := make([]string, 0, len(row))

	for column, value := range row {
		columns = append(columns, string(column))
		values = append(values, value)
	}

	return fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES (%s);\n", table, strings.Join(columns, ","), strings.Join(values, ","))
}

var DateTime = newStaticColumn("NOW()")

var BigInt = newCallableColumn(func() string {
	return gofakeit.DigitN(uint(gofakeit.Number(0, 10)))
})

var GUID = newCallableColumn(func() string {
	return "'" + gofakeit.UUID() + "'"
})

func String(size int) ColumnGenerator {
	return newCallableColumn(func() string {
		return "'" + gofakeit.LetterN(uint(gofakeit.Number(0, size))) + "'"
	})
}

var Version = newCallableColumn(func() string {
	return "'" + gofakeit.Generate("#.#.#") + "'"
})

var TinyInt = newCallableColumn(func() string {
	return gofakeit.DigitN(uint(gofakeit.Number(0, 1)))
})
