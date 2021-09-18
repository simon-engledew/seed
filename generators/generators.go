package generators

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/types"
	"github.com/simon-engledew/seed/quote"
	"strconv"
)

type ValueGenerator interface {
	Value(ctx context.Context) string
}

type primaryColumn struct {
	count uint64
}

func (c *primaryColumn) Value(ctx context.Context) string {
	c.count += 1
	return strconv.FormatUint(c.count, 10)
}

func Get(ft *types.FieldType, isPrimary bool) ValueGenerator {
	name := types.TypeToStr(ft.Tp, ft.Charset)
	_ = mysql.HasUnsignedFlag(ft.Flag)
	length, _ := mysql.GetDefaultFieldLengthAndDecimal(ft.Tp)

	if isPrimary {
		return &primaryColumn{}
	}

	switch name {
	case "tinyint":
		return &funcGenerator{
			value: func() string {
				return gofakeit.DigitN(uint(gofakeit.Number(0, 1)))
			},
		}
	case "int":
		return &funcGenerator{
			value: func() string {
				return gofakeit.DigitN(uint(gofakeit.Number(0, length)))
			},
		}
	case "datetime":
		return &funcGenerator{
			value: func() string {
				d := gofakeit.Date()
				return quote.Quote(d.Format("2006-01-02 15:04:05"))
			},
		}
	case "bigint":
		return &funcGenerator{
			value: func() string {
				return gofakeit.DigitN(uint(gofakeit.Number(0, length)))
			},
		}
	case "varchar":
		return &funcGenerator{
			value: func() string {
				n := uint(gofakeit.Number(0, length))

				return quote.Quote(gofakeit.LetterN(n))
			},
		}
	case "text":
		return &funcGenerator{
			value: func() string {
				return quote.Quote(gofakeit.HackerPhrase())
			},
		}
	}

	return Identity(ft.InfoSchemaStr())
}
