package generators

import "context"

type funcGenerator struct {
	value func(ctx context.Context) string
}

func (c *funcGenerator) Value(ctx context.Context) string {
	return c.value(ctx)
}

func Func(fn func(ctx context.Context) string) ColumnGenerator {
	return &funcGenerator{
		value: fn,
	}
}
