package generators

import "context"

type funcGenerator struct {
	value func(ctx context.Context) *Value
}

func (c *funcGenerator) Value(ctx context.Context) *Value {
	return c.value(ctx)
}

func Func(fn func(ctx context.Context) *Value) ValueGenerator {
	return &funcGenerator{
		value: fn,
	}
}
