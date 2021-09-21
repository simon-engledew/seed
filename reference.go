package seed

import (
	"context"
	"github.com/simon-engledew/seed/generators"
)

var parentKey contextKey = "parent"

type References interface {
	WithParents(ctx context.Context, current string, rows Rows) context.Context
}

type dependentColumn struct {
	tableName   string
	columnIndex int
}

//func (c *dependentColumn) WithParents(ctx context.Context, current string, rows Rows) context.Context {
//	parents, ok := rows[c.tableName]
//	if !ok {
//		panic(fmt.Errorf("parent missing: %s -> %s", current, c.tableName))
//	}
//	return context.WithValue(ctx, parentKey, parents)
//}

func (c *dependentColumn) Value(ctx context.Context) string {
	parent := ctx.Value(parentKey).(Rows)
	return parent[c.tableName][c.columnIndex]
}

func Reference(tableName string, columnIndex int) generators.ValueGenerator {
	return &dependentColumn{
		tableName:   tableName,
		columnIndex: columnIndex,
	}
}
