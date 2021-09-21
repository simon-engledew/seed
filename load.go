package seed

import (
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"strconv"
	"strings"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/jinzhu/inflection"
	"github.com/jpillora/longestcommon"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/types"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"github.com/simon-engledew/seed/escape"
	"github.com/simon-engledew/seed/generators"
)

type Row []string
type Rows map[string]Row
type Column struct {
	Name      string
	Generator generators.ValueGenerator
}

func generator(ft *types.FieldType, isPrimary bool) generators.ValueGenerator {
	if isPrimary {
		return generators.Counter()
	}

	name := types.TypeToStr(ft.Tp, ft.Charset)
	_ = mysql.HasUnsignedFlag(ft.Flag)
	length := ft.Flen
	if length == types.UnspecifiedLength {
		length, _ = mysql.GetDefaultFieldLengthAndDecimal(ft.Tp)
	}

	switch name {
	case "tinyint":
		return generators.Format("{number:0,1}")
	case "smallint", "int", "bigint":
		return generators.Faker(func(f *gofakeit.Faker) string {
			return f.DigitN(uint(f.Number(0, length)))
		})
	case "double":
		return generators.Faker(func(f *gofakeit.Faker) string {
			return strconv.FormatFloat(f.Float64Range(-100, 100), 'f', -1, 64)
		})
	case "datetime":
		return generators.Faker(func(f *gofakeit.Faker) string {
			return escape.Quote(f.Date().Format("2006-01-02 15:04:05"))
		})
	case "varchar", "varbinary":
		return generators.Faker(func(f *gofakeit.Faker) string {
			n := uint(math.Floor(math.Pow(f.Rand.Float64(), 4) * (1 + float64(length))))
			return escape.Quote(f.LetterN(n))
		})
	case "binary":
		return generators.Faker(func(f *gofakeit.Faker) string {
			return escape.Quote(f.LetterN(uint(length)))
		})
	case "json":
		return generators.Identity(escape.Quote("{}"))
	case "mediumtext", "text":
		return generators.Faker(func(f *gofakeit.Faker) string {
			return escape.Quote(f.HackerPhrase())
		})
	}

	return generators.Identity(escape.Quote(ft.InfoSchemaStr()))
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
					Generator: generator(col.Tp, isPrimary),
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
