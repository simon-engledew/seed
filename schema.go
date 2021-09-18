package seed

import (
	"github.com/simon-engledew/seed/generators"
	"io"
)

type Schema map[TableName]map[ColumnName]generators.ValueGenerator

func (s Schema) Generator(w io.Writer) Generator {
	return &insertStack{
		w:      w,
		schema: s,
		stack:  make(Rows),
	}
}

type SchemaTransform func(t TableName, c ColumnName, g generators.ValueGenerator) generators.ValueGenerator

func (s Schema) Transform(transform SchemaTransform) {
	for t, columns := range s {
		for c, g := range columns {
			columns[c] = transform(t, c, g)
		}
	}
}

func (s Schema) Merge(other Schema) {
	for t, columns := range other {
		for c, g := range columns {
			s[t][c] = g
		}
	}
}
