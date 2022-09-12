package generators

import (
	"context"
	"github.com/simon-engledew/seed/consumers"
	"sync"
)

var rowKey contextKey = "row"
var columnsKey contextKey = "columns"

// WithSiblings stores sibling columns in the context for use in lazy generators.
func WithSiblings(ctx context.Context, row *consumers.Row) context.Context {
	return context.WithValue(ctx, rowKey, row)
}

func WithColumns(ctx context.Context, columns []string) context.Context {
	return context.WithValue(ctx, columnsKey, columns)
}

type lazyGenerator struct {
	fn func(context.Context, map[string]consumers.Value) consumers.Value
}

type lazyValue struct {
	parent *lazyGenerator
	value  consumers.Value
	ctx    context.Context
	once   sync.Once
}

func (v *lazyValue) Value() consumers.Value {
	v.once.Do(func() {
		row := *v.ctx.Value(rowKey).(*consumers.Row)
		columns := v.ctx.Value(columnsKey).([]string)

		mapped := make(map[string]consumers.Value, len(columns))
		for n, column := range columns {
			mapped[column] = row[n]
		}

		v.value = v.parent.fn(v.ctx, mapped)
	})
	return v.value
}

func (v *lazyValue) String() string {
	return v.Value().String()
}

func (v *lazyValue) Escape() bool {
	return v.Value().Escape()
}

func (g *lazyGenerator) Value(ctx context.Context) consumers.Value {
	return &lazyValue{
		parent: g,
		ctx:    ctx,
	}
}

func Lazy(fn func(ctx context.Context, row map[string]consumers.Value) consumers.Value) consumers.ValueGenerator {
	return &lazyGenerator{
		fn: fn,
	}
}
