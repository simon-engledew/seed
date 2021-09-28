package generators

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed/escape"
	"strconv"
	"sync"
)

type ColumnGenerator interface {
	Value(ctx context.Context) string
}

func Counter() ColumnGenerator {
	c := uint64(0)
	return Locked(func() string {
		c += 1
		return strconv.FormatUint(c, 10)
	})
}

func Locked(fn func() string) ColumnGenerator {
	var m sync.Mutex
	return Func(func(ctx context.Context) string {
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

func Faker(fn func(*gofakeit.Faker) string) ColumnGenerator {
	return Locked(func() string {
		f := fakers.Get().(*gofakeit.Faker)
		defer fakers.Put(f)
		return fn(f)
	})
}

func Format(fmt string) ColumnGenerator {
	return Faker(func(f *gofakeit.Faker) string {
		return escape.Quote(f.Generate(fmt))
	})
}
