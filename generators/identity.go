package generators

import (
	"context"
)

type identityGenerator struct {
	value string
}

func (c *identityGenerator) String() string {
	return "<" + c.value + ">"
}

func (c *identityGenerator) Value(ctx context.Context) string {
	return c.value
}

func Identity(value string) ValueGenerator {
	return &identityGenerator{
		value: value,
	}
}
