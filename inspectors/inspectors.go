package inspectors

import "github.com/shaaraddalvi/seed/generators"

// Inspector returns a function that can provide metadata about the schema we are building test data for.
type Inspector func(func(tableName, columnName string, column ColumnInfo)) error

type ColumnInfo interface {
	Generator() generators.ValueGenerator
	Type() string
}
