package generators

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"strconv"
	"sync"
)

type Value struct {
	Value string
	Quote bool
}

func NewValue(v string, q bool) *Value {
	return &Value{Value: v, Quote: q}
}

type ValueGenerator interface {
	Value(ctx context.Context) *Value
}

func Counter() ValueGenerator {
	c := uint64(0)
	return Locked(func() *Value {
		c += 1
		return NewValue(strconv.FormatUint(c, 10), false)
	})
}

func Locked(fn func() *Value) ValueGenerator {
	var m sync.Mutex
	return Func(func(ctx context.Context) *Value {
		m.Lock()
		defer m.Unlock()
		return fn()
	})
}

var fakers = sync.Pool{
	New: func() interface{} {
		return gofakeit.New(0)
	},
}

func Faker(fn func(*gofakeit.Faker) (string, bool)) ValueGenerator {
	return Func(func(ctx context.Context) *Value {
		f := fakers.Get().(*gofakeit.Faker)
		defer fakers.Put(f)
		return NewValue(fn(f))
	})
}

func Format(fmt string) ValueGenerator {
	return Faker(func(f *gofakeit.Faker) (string, bool) {
		return f.Generate(fmt), true
	})
}
