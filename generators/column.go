package generators

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed/consumers"
	"math"
)

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
