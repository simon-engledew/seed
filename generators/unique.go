package generators

import (
	"context"
	"fmt"
	"github.com/simon-engledew/seed/consumers"
	"strings"
)

func Unique(generator consumers.ValueGenerator) consumers.ValueGenerator {
	seen := make(map[string]struct{})
	return Locked(Func(func(ctx context.Context) consumers.Value {
		for {
			v := generator.Value(ctx)
			key := fmt.Sprintf("%t:%q", v.Escape(), v.String())
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				return v
			}
		}
	}))
}

func UniqueRow(generator consumers.ValueGenerator, columns ...string) consumers.ValueGenerator {
	seen := make(map[string]struct{})
	return Locked(Lazy(func(ctx context.Context, row map[string]consumers.Value) consumers.Value {
		for {
			v := generator.Value(ctx)

			var buf strings.Builder

			_, _ = fmt.Fprintf(&buf, "%t:%q", v.Escape(), v.String())

			for _, column := range columns {
				c := row[column]
				_, _ = fmt.Fprintf(&buf, ",%t:%q", c.Escape(), c.String())
			}

			key := buf.String()
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				return v
			}
		}
	}))
}
