package seed

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed/distribution"
	"io"
	"sort"
	"strconv"
	"strings"
)

type ColumnGenerator func() string

type TableGenerator func(w io.Writer, tableName string) error

func Generate(schema map[string]TableGenerator) func(io.Writer) error {
	return func(w io.Writer) error {
		gofakeit.Seed(0)

		for tableName, generator := range schema {
			err := generator(w, tableName)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func Insert(dist distribution.Distribution, columns map[string]ColumnGenerator, next ...func(Reference) map[string]TableGenerator) TableGenerator {
	return func(w io.Writer, tableName string) error {
		columnNames := make([]string, 0, len(columns))
		for columnName := range columns {
			columnNames = append(columnNames, columnName)
		}

		sort.Strings(columnNames)

		count := dist()

		for i := uint(0); i < count; i++ {
			mapped := make(map[string]string, len(columns))

			values := make([]string, 0, len(columnNames))
			for _, columnName := range columnNames {
				column := columns[columnName]
				value := column()
				mapped[columnName] = value
				values = append(values, value)
			}

			ref := func(columnName string) ColumnGenerator {
				return Identity(mapped[columnName])
			}

			_, err := fmt.Fprintf(w, "INSERT INTO %s (%s) VALUES (%s);\n", tableName, strings.Join(columnNames, ","), strings.Join(values, ","))
			if err != nil {
				return err
			}

			for _, fn := range next {
				err := Generate(fn(ref))(w)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}
}

func PrimaryKey() ColumnGenerator {
	var id uint64
	return func() string {
		id += 1
		return strconv.FormatUint(id, 10)
	}
}

func Identity(val string) ColumnGenerator {
	return func() string {
		return val
	}
}

func DateTime() string {
	return "NOW()"
}

func BigInt() string {
	return gofakeit.DigitN(uint(gofakeit.Number(0, 10)))
}

func GUID() string {
	return "'" + gofakeit.UUID() + "'"
}

func String(size int) ColumnGenerator {
	return func() string {
		return "'" + gofakeit.LetterN(uint(gofakeit.Number(0, size))) + "'"
	}
}

func Version() string {
	return "'" + gofakeit.Generate("#.#.#") + "'"
}

func TinyInt() string {
	return gofakeit.DigitN(uint(gofakeit.Number(0, 10)))
}

type Reference func(columnName string) ColumnGenerator
