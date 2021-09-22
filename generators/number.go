package generators

import (
	"github.com/brianvoe/gofakeit/v6"
	"strconv"
)

func Number(min, max int) ValueGenerator {
	return Faker(func(f *gofakeit.Faker) string {
		return strconv.Itoa(f.Number(min, max))
	})
}
