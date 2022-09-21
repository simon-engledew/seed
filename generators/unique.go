package generators

import (
	"bytes"
	"context"
	"fmt"
	"github.com/simon-engledew/seed/consumers"
	"sync"
)

func notLazy(fn func(ctx context.Context, row map[string]consumers.Value) consumers.Value) consumers.ValueGenerator {
	return Func(func(ctx context.Context) consumers.Value {
		return fn(ctx, nil)
	})
}

var buffers = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func Unique(generator consumers.ValueGenerator, columns ...string) consumers.ValueGenerator {
	seen := make(map[string]struct{})

	fn := Lazy
	if len(columns) == 0 {
		fn = notLazy
	}

	return Locked(fn(func(ctx context.Context, row map[string]consumers.Value) consumers.Value {
		buf := buffers.Get().(*bytes.Buffer)
		defer buffers.Put(buf)
		for {
			buf.Reset()

			v := generator.Value(ctx)

			_, _ = fmt.Fprintf(buf, "%t:%q", v.Escape(), v.String())

			for _, column := range columns {
				c := row[column]
				_, _ = fmt.Fprintf(buf, ",%t:%q", c.Escape(), c.String())
			}

			key := buf.String()
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				return v
			}
		}
	}))
}
