package generators

import (
	"context"
	"strconv"
)

type ValueGenerator interface {
	Value(ctx context.Context) string
}

type counter struct {
	count uint64
}

func (c *counter) Value(ctx context.Context) string {
	c.count += 1
	return strconv.FormatUint(c.count, 10)
}

func Counter() ValueGenerator {
	return &counter{}
}
