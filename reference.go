package seed

import (
	"context"
	"github.com/simon-engledew/seed/generators"
)

var parentKey contextKey = "parent"

type References interface {
	WithParents(ctx context.Context, rows Rows) context.Context
}

type dependentColumn struct {
	tableName  TableName
	columnName ColumnName
}

func (c *dependentColumn) WithParents(ctx context.Context, rows Rows) context.Context {
	return context.WithValue(ctx, parentKey, rows[c.tableName])
}

func (c *dependentColumn) Value(ctx context.Context) string {
	parent := ctx.Value(parentKey).(Row)
	return parent[c.columnName]
}

func Reference(tableName TableName, columnName ColumnName) generators.ValueGenerator {
	return &dependentColumn{
		tableName:  tableName,
		columnName: columnName,
	}
}
