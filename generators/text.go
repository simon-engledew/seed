package generators

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed/quote"
)

func Text(size int) ValueGenerator {
	return &funcGenerator{
		value: func() string {
			n := uint(gofakeit.Number(0, size))

			return quote.Quote(gofakeit.LetterN(n))
		},
	}
}

func TextFormat(fmt string) ValueGenerator {
	return &funcGenerator{
		value: func() string {
			return quote.Quote(gofakeit.Generate(fmt))
		},
	}
}
