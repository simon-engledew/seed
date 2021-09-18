package generators

import "context"

type funcGenerator struct {
	value func() string
}

func (c *funcGenerator) Value(ctx context.Context) string {
	return c.value()
}
