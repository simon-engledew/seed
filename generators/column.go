package generators

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/types"
	"github.com/simon-engledew/seed/escape"
	"math"
	"strconv"
)

func formatUint(fn func(*gofakeit.Faker) uint64) ColumnGenerator {
	f := gofakeit.New(0)
	return Locked(func() string {
		return strconv.FormatUint(fn(f), 10)
	})
}

func formatInt(fn func(*gofakeit.Faker) int64) ColumnGenerator {
	f := gofakeit.New(0)
	return Locked(func() string {
		return strconv.FormatInt(fn(f), 10)
	})
}

func Column(ft *types.FieldType, isPrimary bool) ColumnGenerator {
	if isPrimary {
		return Counter()
	}

	name := types.TypeToStr(ft.Tp, ft.Charset)
	isUnsigned := mysql.HasUnsignedFlag(ft.Flag)
	length := ft.Flen
	if length == types.UnspecifiedLength {
		length, _ = mysql.GetDefaultFieldLengthAndDecimal(ft.Tp)
	}

	switch name {
	case "tinyint":
		if isUnsigned {
			return formatUint(func(f *gofakeit.Faker) uint64 {
				return uint64(f.Uint8())
			})
		}
		return formatInt(func(f *gofakeit.Faker) int64 {
			return int64(f.Int8())
		})
	case "smallint":
		if isUnsigned {
			return formatUint(func(f *gofakeit.Faker) uint64 {
				return uint64(f.Uint16())
			})
		}
		return formatInt(func(f *gofakeit.Faker) int64 {
			return int64(f.Int16())
		})
	case "int":
		if isUnsigned {
			return formatUint(func(f *gofakeit.Faker) uint64 {
				return uint64(f.Uint32())
			})
		}
		return formatInt(func(f *gofakeit.Faker) int64 {
			return int64(f.Int32())
		})
	case "bigint":
		if isUnsigned {
			return formatUint(func(f *gofakeit.Faker) uint64 {
				return f.Uint64()
			})
		}
		return formatInt(func(f *gofakeit.Faker) int64 {
			return f.Int64()
		})
	case "double":
		return Faker(func(f *gofakeit.Faker) string {
			return strconv.FormatFloat(f.Float64Range(-100, 100), 'f', -1, 64)
		})
	case "datetime":
		return Faker(func(f *gofakeit.Faker) string {
			return escape.Quote(f.Date().Format("2006-01-02 15:04:05"))
		})
	case "varchar", "varbinary":
		return Faker(func(f *gofakeit.Faker) string {
			n := uint(math.Floor(math.Pow(f.Rand.Float64(), 4) * (1 + float64(length))))
			return escape.Quote(f.LetterN(n))
		})
	case "binary":
		return Faker(func(f *gofakeit.Faker) string {
			return escape.Quote(f.LetterN(uint(length)))
		})
	case "json":
		return Identity(escape.Quote("{}"))
	case "mediumtext", "text":
		return Faker(func(f *gofakeit.Faker) string {
			return escape.Quote(f.HackerPhrase())
		})
	}

	return Identity(escape.Quote(ft.InfoSchemaStr()))
}
