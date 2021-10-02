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
        JSON_ARRAYAGG(,
			JSON_OBJECT(
			    'name', COLUMN_NAME,
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
	Name       string `json:"name"`
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
			return generators.Faker(func(f *gofakeit.Faker) *generators.Value {
				return generators.NewValue(strconv.FormatUint(uint64(f.Uint8()), 10), false)
			})
		}
		return generators.Faker(func(f *gofakeit.Faker) *generators.Value {
			return generators.NewValue(strconv.FormatInt(int64(f.Int8()), 10), false)
		})
	case "smallint":
		if c.IsUnsigned {
			return generators.Faker(func(f *gofakeit.Faker) *generators.Value {
				return generators.NewValue(strconv.FormatUint(uint64(f.Uint16()), 10), false)
			})
		}
		return generators.Faker(func(f *gofakeit.Faker) *generators.Value {
			return generators.NewValue(strconv.FormatInt(int64(f.Int16()), 10), false)
		})
	case "int":
		if c.IsUnsigned {
			return generators.Faker(func(f *gofakeit.Faker) *generators.Value {
				return generators.NewValue(strconv.FormatUint(uint64(f.Uint32()), 10), false)
			})
		}
		return generators.Faker(func(f *gofakeit.Faker) *generators.Value {
			return generators.NewValue(strconv.FormatInt(int64(f.Int32()), 10), false)
		})
	case "bigint":
		if c.IsUnsigned {
			return generators.Faker(func(f *gofakeit.Faker) *generators.Value {
				return generators.NewValue(strconv.FormatUint(f.Uint64(), 10), false)
			})
		}
		return generators.Faker(func(f *gofakeit.Faker) *generators.Value {
			return generators.NewValue(strconv.FormatInt(f.Int64(), 10), false)
		})
	case "double":
		return generators.Faker(func(f *gofakeit.Faker) *generators.Value {
			return generators.NewValue(strconv.FormatFloat(f.Float64Range(-100, 100), 'f', -1, 64), false)
		})
	case "datetime":
		return generators.Faker(func(f *gofakeit.Faker) *generators.Value {
			return generators.NewValue(f.Date().Format("'2006-01-02 15:04:05'"), false)
		})
	case "varchar", "varbinary":
		return generators.Faker(func(f *gofakeit.Faker) *generators.Value {
			n := uint(math.Floor(math.Pow(f.Rand.Float64(), 4) * (1 + float64(c.Length))))
			return generators.NewValue(f.LetterN(n), true)
		})
	case "binary":
		return generators.Faker(func(f *gofakeit.Faker) *generators.Value {
			return generators.NewValue(f.LetterN(uint(c.Length)), true)
		})
	case "json":
		return generators.Identity("'{}'", false)
	case "mediumtext", "text":
		return generators.Faker(func(f *gofakeit.Faker) *generators.Value {
			return generators.NewValue(f.HackerPhrase(), true)
		})
	}

	return generators.Identity(c.ColumnType, true)
}

// InspectMySQLConnection will select information from information_schema based on the current database.
func InspectMySQLConnection(db *sql.DB) Inspector {
	return func(fn func(tableName, columnName string, column ColumnInfo)) error {
		var data json.RawMessage

		if err := db.QueryRow(query).Scan(&data); err != nil {
			return err
		}

		var out map[string][]*MySQLColumn

		if err := json.Unmarshal(data, &out); err != nil {
			return err
		}

		for tableName, columns := range out {
			for _, column := range columns {
				fn(tableName, column.Name, column)
			}
		}

		return nil
	}
}

func InspectMySQLSchema(r io.Reader) Inspector {
	p := parser.New()

	return func(fn func(tableName, columnName string, column ColumnInfo)) error {
		dump, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}

		statements, _, err := p.Parse(string(dump), "", "")
		if err != nil {
			return fmt.Errorf("failed to parse sqldump: %w", err)
		}

		for _, statement := range statements {
			if create, ok := statement.(*ast.CreateTableStmt); ok {
				tableName := create.Table.Name.String()

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

					fn(tableName, columnName, &MySQLColumn{
						Name:       columnName,
						IsPrimary:  isPrimary,
						IsUnsigned: mysql.HasUnsignedFlag(col.Tp.Flag),
						ColumnType: col.Tp.CompactStr(),
						DataType:   types.TypeToStr(col.Tp.Tp, col.Tp.Charset),
						Length:     length,
					})
				}
			}
		}

		return nil
	}
}
