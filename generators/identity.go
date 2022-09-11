package generators

import (
	"context"
	"fmt"
	"github.com/simon-engledew/seed/consumers"
)

type identityGenerator struct {
	value consumers.Value
}

func (c *identityGenerator) String() string {
	return fmt.Sprintf("<%s>", c.value)
}

func (c *identityGenerator) Value(ctx context.Context) consumers.Value {
	return c.value
}

func Identity(v consumers.Value) consumers.ValueGenerator {
	return &identityGenerator{
		value: v,
	}
}
