package generators

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed/consumers"
	"sync"
)

func Counter() consumers.ValueGenerator {
	c := uint64(0)
	return Locked(Func(func(_ context.Context) consumers.Value {
		c += 1
		return UnsignedInt(c)
	}))
}

func Locked(gen consumers.ValueGenerator) consumers.ValueGenerator {
	var m sync.Mutex
	return Func(func(ctx context.Context) consumers.Value {
		m.Lock()
		defer m.Unlock()
		return gen.Value(ctx)
	})
}

var fakers = sync.Pool{
	New: func() interface{} {
		return gofakeit.New(0)
	},
}

func Faker(fn func(*gofakeit.Faker) consumers.Value) consumers.ValueGenerator {
	return Func(func(ctx context.Context) consumers.Value {
		f := fakers.Get().(*gofakeit.Faker)
		defer fakers.Put(f)
		return fn(f)
	})
}

func Format[T interface {
	consumers.Value
	Quoted | Unquoted
}](fmt string) consumers.ValueGenerator {
	return Faker(func(f *gofakeit.Faker) consumers.Value {
		return T(f.Generate(fmt))
	})
}
