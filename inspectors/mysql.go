package inspectors

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/types"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"github.com/simon-engledew/seed/escape"
	"github.com/simon-engledew/seed/generators"
	"io"
	"io/ioutil"
	"math"
	"strconv"
)

const query = `
SELECT JSON_OBJECTAGG(table_name, columns) FROM (
    SELECT 
        TABLE_NAME AS 'table_name',
        JSON_OBJECTAGG(COLUMN_NAME,
			JSON_OBJECT(
				'data_type', DATA_TYPE,
				'is_primary', COLUMN_KEY = 'PRI',
				'is_unsigned', COLUMN_TYPE LIKE '% unsigned',
			    'column_type', COLUMN_TYPE,
			    'length', CHARACTER_MAXIMUM_LENGTH
			)
		) AS 'columns'
	FROM information_schema.COLUMNS
	WHERE TABLE_SCHEMA = DATABASE()
    AND EXTRA NOT LIKE '% GENERATED'
	GROUP BY TABLE_NAME
) AS pairs`

type MySQLColumn struct {
	DataType   string `json:"data_type"`
	IsPrimary  bool   `json:"is_primary"`
	IsUnsigned bool   `json:"is_unsigned"`
	Length     int    `json:"length"`
	ColumnType string `json:"column_type"`
}

func (c *MySQLColumn) Type() string {
	return c.ColumnType
}

func (c *MySQLColumn) Generator() generators.ValueGenerator {
	if c.IsPrimary {
		return generators.Counter()
	}

	switch c.DataType {
	case "tinyint":
		if c.IsUnsigned {
			return generators.Faker(func(f *gofakeit.Faker) string {
				return strconv.FormatUint(uint64(f.Uint8()), 10)
			})
		}
		return generators.Faker(func(f *gofakeit.Faker) string {
			return strconv.FormatInt(int64(f.Int8()), 10)
		})
	case "smallint":
		if c.IsUnsigned {
			return generators.Faker(func(f *gofakeit.Faker) string {
				return strconv.FormatUint(uint64(f.Uint16()), 10)
			})
		}
		return generators.Faker(func(f *gofakeit.Faker) string {
			return strconv.FormatInt(int64(f.Int16()), 10)
		})
	case "int":
		if c.IsUnsigned {
			return generators.Faker(func(f *gofakeit.Faker) string {
				return strconv.FormatUint(uint64(f.Uint32()), 10)
			})
		}
		return generators.Faker(func(f *gofakeit.Faker) string {
			return strconv.FormatInt(int64(f.Int32()), 10)
		})
	case "bigint":
		if c.IsUnsigned {
			return generators.Faker(func(f *gofakeit.Faker) string {
				return strconv.FormatUint(f.Uint64(), 10)
			})
		}
		return generators.Faker(func(f *gofakeit.Faker) string {
			return strconv.FormatInt(f.Int64(), 10)
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
			n := uint(math.Floor(math.Pow(f.Rand.Float64(), 4) * (1 + float64(c.Length))))
			return escape.Quote(f.LetterN(n))
		})
	case "binary":
		return generators.Faker(func(f *gofakeit.Faker) string {
			return escape.Quote(f.LetterN(uint(c.Length)))
		})
	case "json":
		return generators.Identity(escape.Quote("{}"))
	case "mediumtext", "text":
		return generators.Faker(func(f *gofakeit.Faker) string {
			return escape.Quote(f.HackerPhrase())
		})
	}

	return generators.Identity(escape.Quote(c.ColumnType))
}

// InspectMySQLConnection will select information from information_schema based on the current database.
func InspectMySQLConnection(db *sql.DB) Inspector {
	return func() (map[string]map[string]Column, error) {
		var data json.RawMessage

		if err := db.QueryRow(query).Scan(&data); err != nil {
			return nil, err
		}

		var out map[string]map[string]Column

		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}

		return out, nil
	}
}

func InspectMySQLSchema(r io.Reader) Inspector {
	p := parser.New()

	return func() (map[string]map[string]Column, error) {
		dump, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}

		statements, _, err := p.Parse(string(dump), "", "")
		if err != nil {
			return nil, fmt.Errorf("failed to parse sqldump: %w", err)
		}

		out := make(map[string]map[string]Column)

		var tableNames []string

		for _, statement := range statements {
			if create, ok := statement.(*ast.CreateTableStmt); ok {
				table := make(map[string]Column)

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

			outer:
				for _, col := range create.Cols {
					columnName := col.Name.String()

					for _, option := range col.Options {
						if option.Tp == ast.ColumnOptionGenerated {
							continue outer
						}
						if option.Tp == ast.ColumnOptionPrimaryKey {
							primaryKey[columnName] = struct{}{}
						}
					}

					_, isPrimary := primaryKey[columnName]

					length := col.Tp.Flen
					if length == types.UnspecifiedLength {
						length, _ = mysql.GetDefaultFieldLengthAndDecimal(col.Tp.Tp)
					}

					table[columnName] = &MySQLColumn{
						IsPrimary:  isPrimary,
						IsUnsigned: mysql.HasUnsignedFlag(col.Tp.Flag),
						ColumnType: col.Tp.CompactStr(),
						DataType:   types.TypeToStr(col.Tp.Tp, col.Tp.Charset),
						Length:     length,
					}
				}

				out[tableName] = table
			}
		}

		return out, nil
	}
}
