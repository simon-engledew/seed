package mysql

import (
	"database/sql"
	"encoding/json"
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/generators"
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
	ColumnName string `json:"name"`
	DataType   string `json:"data_type"`
	IsPrimary  bool   `json:"is_primary"`
	IsUnsigned bool   `json:"is_unsigned"`
	Length     int    `json:"length"`
	ColumnType string `json:"column_type"`
}

func (c MySQLColumn) Name() string {
	return c.ColumnName
}

func (c MySQLColumn) Type() string {
	return c.ColumnType
}

func (c MySQLColumn) Generator() generators.ValueGenerator {
	return generators.Column(c.DataType, c.IsUnsigned, c.IsPrimary, c.Length, generators.Identity(c.ColumnType, true))
}

// InspectMySQLConnection will select information from information_schema based on the current database.
func InspectMySQLConnection(db *sql.DB) (map[string][]seed.Column, error) {
	var data json.RawMessage

	if err := db.QueryRow(query).Scan(&data); err != nil {
		return nil, err
	}

	var out map[string][]*MySQLColumn

	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}

	tables := make(map[string][]seed.Column)

	for tableName, columns := range out {
		tables[tableName] = make([]seed.Column, 0, len(columns))
		for _, column := range columns {
			tables[tableName] = append(tables[tableName], column)
		}
	}

	return tables, nil
}