package generators

import (
	"context"
	"github.com/simon-engledew/seed/consumers"
)

type funcGenerator struct {
	valueFunc func(ctx context.Context) consumers.Value
}

func (c *funcGenerator) Value(ctx context.Context) consumers.Value {
	return c.valueFunc(ctx)
}

func Func(fn func(ctx context.Context) consumers.Value) *funcGenerator {
	return &funcGenerator{
		valueFunc: fn,
	}
}
