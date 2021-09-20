package generators

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"strconv"
	"sync"
)

type ValueGenerator interface {
	Value(ctx context.Context) string
}

func Counter() ValueGenerator {
	c := uint64(0)
	return Locked(func() string {
		c += 1
		return strconv.FormatUint(c, 10)
	})
}

func Locked(fn func() string) ValueGenerator {
	var m sync.Mutex
	return Func(func(ctx context.Context) string {
		m.Lock()
		defer m.Unlock()
		return fn()
	})
}

func Faker(fn func(*gofakeit.Faker) string) ValueGenerator {
	f := gofakeit.New(0)
	return Locked(func() string {
		return fn(f)
	})
}
