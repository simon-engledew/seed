package generators

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed/escape"
)

func Format(fmt string) ValueGenerator {
	return Faker(func(f *gofakeit.Faker) string {
		return escape.Quote(f.Generate(fmt))
	})
}
