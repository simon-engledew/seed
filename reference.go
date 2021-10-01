package seed

import (
	"context"
	"fmt"
	"github.com/simon-engledew/seed/generators"
)

type contextKey string

var parentKey contextKey = "parent"

// WithParents stores parent rows in the context for use in references.
func WithParents(ctx context.Context, rows Rows) context.Context {
	return context.WithValue(ctx, parentKey, rows)
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
