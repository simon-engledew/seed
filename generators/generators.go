package generators

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"math"
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

func fakeUint[T uint8 | uint16 | uint32 | uint64](gen func(*gofakeit.Faker) T) ValueGenerator {
	return Faker(func(f *gofakeit.Faker) (string, bool) {
		return strconv.FormatUint(uint64(gen(f)), 10), false
	})
}

func fakeInt[T int8 | int16 | int32 | int64](gen func(*gofakeit.Faker) T) ValueGenerator {
	return Faker(func(f *gofakeit.Faker) (string, bool) {
		return strconv.FormatInt(int64(gen(f)), 10), false
	})
}

func Column(dataType string, isPrimary, isUnsigned bool, length int, fallback ValueGenerator) ValueGenerator {
	if isPrimary {
		return Counter()
	}

	switch dataType {
	case "tinyint":
		if isUnsigned {
			return fakeUint((*gofakeit.Faker).Uint8)
		}
		return fakeInt((*gofakeit.Faker).Int8)
	case "smallint":
		if isUnsigned {
			return fakeUint((*gofakeit.Faker).Uint16)
		}
		return fakeInt((*gofakeit.Faker).Int16)
	case "int":
		if isUnsigned {
			return fakeUint((*gofakeit.Faker).Uint32)
		}
		return fakeInt((*gofakeit.Faker).Int32)
	case "bigint":
		if isUnsigned {
			return fakeUint((*gofakeit.Faker).Uint64)
		}
		return fakeInt((*gofakeit.Faker).Int64)
	case "double":
		return Faker(func(f *gofakeit.Faker) (string, bool) {
			return strconv.FormatFloat(f.Float64Range(-100, 100), 'f', -1, 64), false
		})
	case "datetime":
		return Faker(func(f *gofakeit.Faker) (string, bool) {
			return f.Date().Format("'2006-01-02 15:04:05'"), false
		})
	case "varchar", "varbinary":
		return Faker(func(f *gofakeit.Faker) (string, bool) {
			n := uint(math.Floor(math.Pow(f.Rand.Float64(), 4) * (1 + float64(length))))
			return f.LetterN(n), true
		})
	case "binary":
		return Faker(func(f *gofakeit.Faker) (string, bool) {
			return f.LetterN(uint(length)), true
		})
	case "json":
		return Identity("'{}'", false)
	case "mediumtext", "text":
		return Faker(func(f *gofakeit.Faker) (string, bool) {
			return f.HackerPhrase(), true
		})
	}

	return fallback
}
