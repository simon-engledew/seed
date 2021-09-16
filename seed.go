package seed

import (
	"context"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed/distribution"
	"io"
	"strconv"
	"strings"
)

type ColumnGenerator interface {
	Value(ctx context.Context) string
}

type Dependent interface {
	DependsOn() TableName
}

type Distributed interface {
	Distribution() distribution.Distribution
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
	count        uint64
	distribution distribution.Distribution
}

func (c *primaryColumn) Value(ctx context.Context) string {
	c.count += 1
	return strconv.FormatUint(c.count, 10)
}

func PrimaryKey(dist distribution.Distribution) ColumnGenerator {
	return &primaryColumn{
		distribution: dist,
	}
}

func (c *primaryColumn) Distribution() distribution.Distribution {
	return c.distribution
}

type contextKey string

var parentKey contextKey = "parent"

type dependentColumn struct {
	tableName  TableName
	columnName ColumnName
}

func (c *dependentColumn) DependsOn() TableName {
	return c.tableName
}

func (c *dependentColumn) Value(ctx context.Context) string {
	parent := ctx.Value(parentKey).(map[ColumnName]string)
	return parent[c.columnName]
}

func Reference(tableName TableName, columnName ColumnName) ColumnGenerator {
	return &dependentColumn{
		tableName:  tableName,
		columnName: columnName,
	}
}

func Generate(w io.Writer, schema Schema) error {
	gofakeit.Seed(0)

	graph := make(map[TableName]map[TableName]struct{})
	distributions := make(map[TableName]distribution.Distribution)

	for tableName, columns := range schema {
		graph[tableName] = make(map[TableName]struct{})
		for _, generator := range columns {
			if d, ok := generator.(Dependent); ok {
				parent := d.DependsOn()
				graph[tableName][parent] = struct{}{}
			}
			if d, ok := generator.(Distributed); ok {
				dist := d.Distribution()
				distributions[tableName] = dist
			}
		}
	}

	fmt.Println(graph)

	order := make([][]TableName, 0)

	for len(graph) > 0 {
		var current []TableName
		for table, dependents := range graph {
			if len(dependents) == 0 {
				current = append(current, table)
				delete(graph, table)
			}
		}
		for _, dependents := range graph {
			for _, scheduled := range current {
				delete(dependents, scheduled)
			}
		}
		order = append(order, current)
	}

	for _, stage := range order {
		fmt.Println(stage)
	}

	rows := make(map[TableName]map[ColumnName]string)

	for _, stage := range order {
		for _, table := range stage {
			generators := schema[table]

			row := make(map[ColumnName]string, len(generators))

			rows[table] = row

			for column, generator := range generators {
				ctx := context.Background()

				if d, ok := generator.(Dependent); ok {
					parent := d.DependsOn()

					ctx = context.WithValue(ctx, parentKey, rows[parent])
				}

				row[column] = generator.Value(ctx)
			}

		}
	}

	for table, _ := range schema {
		_, err := insert(w, table, rows[table])
		if err != nil {
			return err
		}
	}

	return nil
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
