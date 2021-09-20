package generators

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed/quote"
)

func Format(fmt string) ValueGenerator {
	return &funcGenerator{
		value: func() string {
			return quote.Quote(gofakeit.Generate(fmt))
		},
	}
}
