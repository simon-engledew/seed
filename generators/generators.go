package generators

import (
	"context"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed/consumers"
	"math"
	"strconv"
	"sync"
)

type Unquoted string

type Quoted string

func (v Unquoted) String() string {
	return string(v)
}

func (v Unquoted) Escape() bool {
	return false
}

func (v Quoted) String() string {
	return string(v)
}

func (v Quoted) Escape() bool {
	return true
}

func Counter() consumers.ValueGenerator {
	c := uint64(0)
	return Locked(func(_ context.Context) consumers.Value {
		c += 1
		return Unquoted(strconv.FormatUint(c, 10))
	})
}

func Locked(fn func(ctx context.Context) consumers.Value) consumers.ValueGenerator {
	var m sync.Mutex
	return Func(func(ctx context.Context) consumers.Value {
		m.Lock()
		defer m.Unlock()
		return fn(ctx)
	})
}

var fakers = sync.Pool{
	New: func() interface{} {
		return gofakeit.New(0)
	},
}

type value interface {
	consumers.Value
	Quoted | Unquoted
}

func Faker[T value](fn func(*gofakeit.Faker) string) consumers.ValueGenerator {
	return Func(func(ctx context.Context) consumers.Value {
		f := fakers.Get().(*gofakeit.Faker)
		defer fakers.Put(f)
		return T(fn(f))
	})
}

func Format[T value](fmt string) consumers.ValueGenerator {
	return Faker[T](func(f *gofakeit.Faker) string {
		return f.Generate(fmt)
	})
}

func fakeUint[T uint8 | uint16 | uint32 | uint64](gen func(*gofakeit.Faker) T) consumers.ValueGenerator {
	return Faker[Unquoted](func(f *gofakeit.Faker) string {
		return strconv.FormatUint(uint64(gen(f)), 10)
	})
}

func fakeInt[T int8 | int16 | int32 | int64](gen func(*gofakeit.Faker) T) consumers.ValueGenerator {
	return Faker[Unquoted](func(f *gofakeit.Faker) string {
		return strconv.FormatInt(int64(gen(f)), 10)
	})
}

func Column(dataType string, isUnsigned bool, length int) consumers.ValueGenerator {
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
		return Faker[Unquoted](func(f *gofakeit.Faker) string {
			return strconv.FormatFloat(f.Float64Range(-100, 100), 'f', -1, 64)
		})
	case "datetime":
		return Faker[Unquoted](func(f *gofakeit.Faker) string {
			return f.Date().Format("'2006-01-02 15:04:05'")
		})
	case "varchar", "varbinary":
		return Faker[Quoted](func(f *gofakeit.Faker) string {
			n := uint(math.Floor(math.Pow(f.Rand.Float64(), 4) * (1 + float64(length))))
			return f.LetterN(n)
		})
	case "binary":
		return Faker[Quoted](func(f *gofakeit.Faker) string {
			return f.LetterN(uint(length))
		})
	case "json":
		return Identity[Unquoted]("'{}'")
	case "mediumtext", "text":
		return Faker[Quoted](func(f *gofakeit.Faker) string {
			return f.HackerPhrase()
		})
	}

	return nil
}

func Unique(generator consumers.ValueGenerator) consumers.ValueGenerator {
	seen := make(map[string]struct{})
	return Locked(func(ctx context.Context) consumers.Value {
		for {
			v := generator.Value(ctx)
			key := fmt.Sprintf("%v:%s", v.Escape(), v.String())
			if _, ok := seen[key]; !ok {
				seen[key] = struct{}{}
				return v
			}
		}
	})
}
