package inspectors

import (
	"database/sql"
	"encoding/json"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed/generators"
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

func (c MySQLColumn) Type() string {
	return c.ColumnType
}

func fakeUint[T uint8 | uint16 | uint32 | uint64](gen func(*gofakeit.Faker) T) generators.ValueGenerator {
	return generators.Faker(func(f *gofakeit.Faker) (string, bool) {
		return strconv.FormatUint(uint64(gen(f)), 10), false
	})
}

func fakeInt[T int8 | int16 | int32 | int64](gen func(*gofakeit.Faker) T) generators.ValueGenerator {
	return generators.Faker(func(f *gofakeit.Faker) (string, bool) {
		return strconv.FormatInt(int64(gen(f)), 10), false
	})
}

func (c MySQLColumn) Generator() generators.ValueGenerator {
	if c.IsPrimary {
		return generators.Counter()
	}

	switch c.DataType {
	case "tinyint":
		if c.IsUnsigned {
			return fakeUint((*gofakeit.Faker).Uint8)
		}
		fakeInt((*gofakeit.Faker).Int8)
	case "smallint":
		if c.IsUnsigned {
			return fakeUint((*gofakeit.Faker).Uint16)
		}
		fakeInt((*gofakeit.Faker).Int16)
	case "int":
		if c.IsUnsigned {
			return fakeUint((*gofakeit.Faker).Uint32)
		}
		fakeInt((*gofakeit.Faker).Int32)
	case "bigint":
		if c.IsUnsigned {
			return fakeUint((*gofakeit.Faker).Uint64)
		}
		fakeInt((*gofakeit.Faker).Int64)
	case "double":
		return generators.Faker(func(f *gofakeit.Faker) (string, bool) {
			return strconv.FormatFloat(f.Float64Range(-100, 100), 'f', -1, 64), false
		})
	case "datetime":
		return generators.Faker(func(f *gofakeit.Faker) (string, bool) {
			return f.Date().Format("'2006-01-02 15:04:05'"), false
		})
	case "varchar", "varbinary":
		return generators.Faker(func(f *gofakeit.Faker) (string, bool) {
			n := uint(math.Floor(math.Pow(f.Rand.Float64(), 4) * (1 + float64(c.Length))))
			return f.LetterN(n), true
		})
	case "binary":
		return generators.Faker(func(f *gofakeit.Faker) (string, bool) {
			return f.LetterN(uint(c.Length)), true
		})
	case "json":
		return generators.Identity("'{}'", false)
	case "mediumtext", "text":
		return generators.Faker(func(f *gofakeit.Faker) (string, bool) {
			return f.HackerPhrase(), true
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
