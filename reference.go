package seed

import (
	"context"
	"fmt"
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
	row, ok := parent[c.tableName]
	if !ok {
		panic(fmt.Errorf("referenced table that has not been generated: %q", c.tableName))
	}
	return row[c.columnIndex]
}

func Reference(tableName string, columnIndex int) generators.ValueGenerator {
	return &dependentColumn{
		tableName:   tableName,
		columnIndex: columnIndex,
	}
}
