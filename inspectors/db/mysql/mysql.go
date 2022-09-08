package mysql

import (
	"database/sql"
	"encoding/json"
)

const query = `
SELECT JSON_OBJECTAGG(table_name, columns) FROM (
    SELECT
        TABLE_NAME AS 'table_name',
        JSON_ARRAYAGG(
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

// InspectMySQLConnection will select information from information_schema based on the current database.
func InspectMySQLConnection(db *sql.DB) ([]byte, error) {
	var data json.RawMessage

	if err := db.QueryRow(query).Scan(&data); err != nil {
		return nil, err
	}

	return data, nil
}
