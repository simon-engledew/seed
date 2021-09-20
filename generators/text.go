package generators

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed/quote"
)

func Format(fmt string) ValueGenerator {
	return Faker(func(f *gofakeit.Faker) string {
		return quote.Quote(f.Generate(fmt))
	})
}
