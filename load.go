package seed

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/jinzhu/inflection"
	"github.com/jpillora/longestcommon"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/types"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"github.com/simon-engledew/seed/generators"
	"github.com/simon-engledew/seed/quote"
	"io"
	"io/ioutil"
	"math/rand"
	"strings"
)

type Row []string
type Rows map[string]Row
type Column struct {
	Name      string
	Generator generators.ValueGenerator
}

func generator(f *gofakeit.Faker, ft *types.FieldType, isPrimary bool) generators.ValueGenerator {
	if isPrimary {
		return generators.Counter()
	}

	name := types.TypeToStr(ft.Tp, ft.Charset)
	_ = mysql.HasUnsignedFlag(ft.Flag)
	length, _ := mysql.GetDefaultFieldLengthAndDecimal(ft.Tp)

	switch name {
	case "tinyint":
		return generators.Format("{number:0,1}")
	case "int":
		return generators.Func(func() string {
			return f.DigitN(uint(f.Number(0, length)))
		})
	case "datetime":
		return generators.Func(func() string {
			return quote.Quote(f.Date().Format("2006-01-02 15:04:05"))
		})
	case "bigint":
		return generators.Func(func() string {
			return f.DigitN(uint(f.Number(0, length)))
		})
	case "varchar":
		return generators.Func(func() string {
			n := uint(f.Number(0, length))

			return quote.Quote(f.LetterN(n))
		})
	case "json":
		return generators.Identity(quote.Quote("{}"))
	case "text":
		return generators.Func(func() string {
			return quote.Quote(f.HackerPhrase())
		})
	}

	return generators.Identity(quote.Quote(ft.InfoSchemaStr()))
}

func Load(r io.Reader) (Schema, error) {
	p := parser.New()

	dump, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	statements, _, err := p.Parse(string(dump), "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to parse sqldump: %w", err)
	}

	schema := make(map[string][]*Column)

	var tableNames []string

	for _, statement := range statements {
		if create, ok := statement.(*ast.CreateTableStmt); ok {
			f := &gofakeit.Faker{Rand: rand.New(rand.NewSource(0))}

			table := make([]*Column, 0, len(create.Cols))

			tableName := create.Table.Name.String()

			tableNames = append(tableNames, tableName)

			primaryKey := make(map[string]struct{})

			for _, constraint := range create.Constraints {
				if constraint.Tp == ast.ConstraintPrimaryKey {
					for _, key := range constraint.Keys {
						columnName := key.Column.String()

						primaryKey[columnName] = struct{}{}
					}
				}
			}

			for _, col := range create.Cols {
				columnName := col.Name.String()

				for _, option := range col.Options {
					if option.Tp == ast.ColumnOptionPrimaryKey {
						primaryKey[columnName] = struct{}{}
					}
				}

				_, isPrimary := primaryKey[columnName]

				table = append(table, &Column{
					Name:      columnName,
					Generator: generator(f, col.Tp, isPrimary),
				})
			}

			schema[tableName] = table
		}
	}

	prefix := longestcommon.Prefix(tableNames)

	for _, columns := range schema {
		for i, column := range columns {
			if strings.HasSuffix(column.Name, "_id") {
				tableName := prefix + inflection.Plural(column.Name[:len(column.Name)-3])

				if parent, ok := schema[tableName]; ok {
					for j, target := range parent {
						if target.Name == "id" {
							columns[i].Generator = Reference(tableName, j)
						}
					}
				}
			}
		}
	}

	return schema, nil
}
