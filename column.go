package seed

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed/escape"
	"github.com/simon-engledew/seed/generators"
	"github.com/simon-engledew/seed/inspectors"
	"math"
	"strconv"
)

func FromColumnInfo(c *inspectors.ColumnInfo) generators.ValueGenerator {
	if c.IsPrimary {
		return generators.Counter()
	}

	switch c.DataType {
	case "tinyint":
		if c.IsUnsigned {
			return generators.Faker(func(f *gofakeit.Faker) string {
				return strconv.FormatUint(uint64(f.Uint8()), 10)
			})
		}
		return generators.Faker(func(f *gofakeit.Faker) string {
			return strconv.FormatInt(int64(f.Int8()), 10)
		})
	case "smallint":
		if c.IsUnsigned {
			return generators.Faker(func(f *gofakeit.Faker) string {
				return strconv.FormatUint(uint64(f.Uint16()), 10)
			})
		}
		return generators.Faker(func(f *gofakeit.Faker) string {
			return strconv.FormatInt(int64(f.Int16()), 10)
		})
	case "int":
		if c.IsUnsigned {
			return generators.Faker(func(f *gofakeit.Faker) string {
				return strconv.FormatUint(uint64(f.Uint32()), 10)
			})
		}
		return generators.Faker(func(f *gofakeit.Faker) string {
			return strconv.FormatInt(int64(f.Int32()), 10)
		})
	case "bigint":
		if c.IsUnsigned {
			return generators.Faker(func(f *gofakeit.Faker) string {
				return strconv.FormatUint(f.Uint64(), 10)
			})
		}
		return generators.Faker(func(f *gofakeit.Faker) string {
			return strconv.FormatInt(f.Int64(), 10)
		})
	case "double":
		return generators.Faker(func(f *gofakeit.Faker) string {
			return strconv.FormatFloat(f.Float64Range(-100, 100), 'f', -1, 64)
		})
	case "datetime":
		return generators.Faker(func(f *gofakeit.Faker) string {
			return escape.Quote(f.Date().Format("2006-01-02 15:04:05"))
		})
	case "varchar", "varbinary":
		return generators.Faker(func(f *gofakeit.Faker) string {
			n := uint(math.Floor(math.Pow(f.Rand.Float64(), 4) * (1 + float64(c.Length))))
			return escape.Quote(f.LetterN(n))
		})
	case "binary":
		return generators.Faker(func(f *gofakeit.Faker) string {
			return escape.Quote(f.LetterN(uint(c.Length)))
		})
	case "json":
		return generators.Identity(escape.Quote("{}"))
	case "mediumtext", "text":
		return generators.Faker(func(f *gofakeit.Faker) string {
			return escape.Quote(f.HackerPhrase())
		})
	}

	return generators.Identity(escape.Quote(c.Type))
}
