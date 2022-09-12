package generators

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed/consumers"
	"math"
	"strconv"
	"sync"
	"time"
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

type Bool bool

func (u Bool) String() string {
	return strconv.FormatBool(bool(u))
}

func (u Bool) Escape() bool {
	return false
}

type Date time.Time

func (u Date) String() string {
	return time.Time(u).Format("'2006-01-02 15:04:05'")
}

func (u Date) Escape() bool {
	return false
}

type Double float64

func (u Double) String() string {
	return strconv.FormatFloat(float64(u), 'f', -1, 64)
}

func (u Double) Escape() bool {
	return false
}

type UnsignedInt uint64

func (u UnsignedInt) String() string {
	return strconv.FormatUint(uint64(u), 10)
}

func (u UnsignedInt) Escape() bool {
	return false
}

type SignedInt int64

func (u SignedInt) String() string {
	return strconv.FormatInt(int64(u), 10)
}

func (u SignedInt) Escape() bool {
	return false
}

func Column(dataType string, isUnsigned bool, length int) consumers.ValueGenerator {
	switch dataType {
	case "tinyint":
		if isUnsigned {
			return Faker(func(f *gofakeit.Faker) consumers.Value {
				return UnsignedInt(f.Uint8())
			})
		}
		return Faker(func(f *gofakeit.Faker) consumers.Value {
			return SignedInt(f.Int8())
		})
	case "smallint":
		if isUnsigned {
			return Faker(func(f *gofakeit.Faker) consumers.Value {
				return UnsignedInt(f.Uint16())
			})
		}
		return Faker(func(f *gofakeit.Faker) consumers.Value {
			return SignedInt(f.Int16())
		})
	case "int":
		if isUnsigned {
			return Faker(func(f *gofakeit.Faker) consumers.Value {
				return UnsignedInt(f.Uint32())
			})
		}
		return Faker(func(f *gofakeit.Faker) consumers.Value {
			return SignedInt(f.Int32())
		})
	case "bigint":
		if isUnsigned {
			return Faker(func(f *gofakeit.Faker) consumers.Value {
				return UnsignedInt(f.Uint64())
			})
		}
		return Faker(func(f *gofakeit.Faker) consumers.Value {
			return SignedInt(f.Int64())
		})
	case "double":
		return Faker(func(f *gofakeit.Faker) consumers.Value {
			return Double(f.Float64Range(-100, 100))
		})
	case "datetime":
		return Faker(func(f *gofakeit.Faker) consumers.Value {
			return Date(f.Date())
		})
	case "varchar", "varbinary":
		return Faker(func(f *gofakeit.Faker) consumers.Value {
			n := uint(math.Floor(math.Pow(f.Rand.Float64(), 4) * (1 + float64(length))))
			return Quoted(f.LetterN(n))
		})
	case "binary":
		return Faker(func(f *gofakeit.Faker) consumers.Value {
			return Quoted(f.LetterN(uint(length)))
		})
	case "json":
		return Identity(Unquoted("'{}'"))
	case "mediumtext", "text":
		return Faker(func(f *gofakeit.Faker) consumers.Value {
			return Quoted(f.HackerPhrase())
		})
	}

	return nil
}
