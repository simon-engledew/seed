package generators

import (
	"context"
	"github.com/simon-engledew/seed/consumers"
	"sync"
)

var rowKey contextKey = "row"

// WithSiblings stores sibling columns in the context for use in lazy generators.
func WithSiblings(ctx context.Context, row *consumers.Row) context.Context {
	return context.WithValue(ctx, rowKey, row)
}

type lazyGenerator struct {
	fn func(row consumers.Row) consumers.Value
}

type lazyValue struct {
	parent *lazyGenerator
	value  consumers.Value
	row    *consumers.Row
	once   sync.Once
}

func (v *lazyValue) Value() consumers.Value {
	v.once.Do(func() {
		v.value = v.parent.fn(*v.row)
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
		row:    ctx.Value(rowKey).(*consumers.Row),
	}
}

func Lazy(fn func(row consumers.Row) consumers.Value) consumers.ValueGenerator {
	return &lazyGenerator{
		fn: fn,
	}
}
