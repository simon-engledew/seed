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
	"strings"
)

type TableName string
type ColumnName string
type Row map[ColumnName]string
type Rows map[TableName]Row

func generator(ft *types.FieldType, isPrimary bool) generators.ValueGenerator {
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
			return gofakeit.DigitN(uint(gofakeit.Number(0, length)))
		})
	case "datetime":
		return generators.Func(func() string {
			return quote.Quote(gofakeit.Date().Format("2006-01-02 15:04:05"))
		})
	case "bigint":
		return generators.Func(func() string {
			return gofakeit.DigitN(uint(gofakeit.Number(0, length)))
		})
	case "varchar":
		return generators.Func(func() string {
			n := uint(gofakeit.Number(0, length))

			return quote.Quote(gofakeit.LetterN(n))
		})
	case "json":
		return generators.Identity(quote.Quote("{}"))
	case "text":
		return generators.Func(func() string {
			return quote.Quote(gofakeit.HackerPhrase())
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

	schema := make(map[TableName]map[ColumnName]generators.ValueGenerator)

	var tableNames []string

	for _, statement := range statements {
		if create, ok := statement.(*ast.CreateTableStmt); ok {
			table := make(map[ColumnName]generators.ValueGenerator)

			tableName := create.Table.Name.String()

			schema[TableName(tableName)] = table

			tableNames = append(tableNames, tableName)

			primaryKey := make(map[ColumnName]struct{})

			for _, constraint := range create.Constraints {
				if constraint.Tp == ast.ConstraintPrimaryKey {
					for _, key := range constraint.Keys {
						columnName := ColumnName(key.Column.String())

						primaryKey[columnName] = struct{}{}
					}
				}
			}

			for _, col := range create.Cols {
				columnName := ColumnName(col.Name.String())

				for _, option := range col.Options {
					if option.Tp == ast.ColumnOptionPrimaryKey {
						primaryKey[columnName] = struct{}{}
					}
				}

				_, isPrimary := primaryKey[columnName]

				table[columnName] = generator(col.Tp, isPrimary)
			}
		}
	}

	prefix := longestcommon.Prefix(tableNames)

	for _, columns := range schema {
		for columnName := range columns {
			column := string(columnName)
			if strings.HasSuffix(column, "_id") {
				tableName := TableName(prefix + inflection.Plural(column[:len(columnName)-3]))
				if _, ok := schema[tableName]; ok {
					columns[columnName] = Reference(tableName, "id")
				}
			}
		}
	}

	return schema, nil
}
