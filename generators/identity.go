package generators

import (
	"context"
	"fmt"
)

type identityGenerator struct {
	value *Value
}

func (c *identityGenerator) String() string {
	return fmt.Sprintf("<%s>", c.value.Value)
}

func (c *identityGenerator) Value(ctx context.Context) *Value {
	return c.value
}

func Identity(value string, quote bool) ValueGenerator {
	return &identityGenerator{
		value: NewValue(value, quote),
	}
}
