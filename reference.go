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
