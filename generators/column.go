package generators

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/types"
	"github.com/simon-engledew/seed/escape"
	"math"
	"strconv"
)

func Column(ft *types.FieldType, isPrimary bool) ValueGenerator {
	if isPrimary {
		return Counter()
	}

	name := types.TypeToStr(ft.Tp, ft.Charset)
	_ = mysql.HasUnsignedFlag(ft.Flag)
	length := ft.Flen
	if length == types.UnspecifiedLength {
		length, _ = mysql.GetDefaultFieldLengthAndDecimal(ft.Tp)
	}

	switch name {
	case "tinyint":
		return Format("{number:0,1}")
	case "smallint", "int", "bigint":
		return Faker(func(f *gofakeit.Faker) string {
			return f.DigitN(uint(f.Number(0, length)))
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
