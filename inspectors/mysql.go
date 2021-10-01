package inspectors

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/types"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"io"
	"io/ioutil"
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
func InspectMySQLConnection(db *sql.DB) Inspector {
	return func() (map[string][]ColumnInfo, error) {
		var data json.RawMessage

		if err := db.QueryRow(query).Scan(&data); err != nil {
			return nil, err
		}

		var out map[string][]ColumnInfo

		if err := json.Unmarshal(data, &out); err != nil {
			return nil, err
		}

		return out, nil
	}
}

func InspectMySQLSchema(r io.Reader) Inspector {
	p := parser.New()

	return func() (map[string][]ColumnInfo, error) {
		dump, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}

		statements, _, err := p.Parse(string(dump), "", "")
		if err != nil {
			return nil, fmt.Errorf("failed to parse sqldump: %w", err)
		}

		out := make(map[string][]ColumnInfo)

		var tableNames []string

		for _, statement := range statements {
			if create, ok := statement.(*ast.CreateTableStmt); ok {
				table := make([]ColumnInfo, 0, len(create.Cols))

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

					table = append(table, ColumnInfo{
						Name:       columnName,
						IsPrimary:  isPrimary,
						IsUnsigned: mysql.HasUnsignedFlag(col.Tp.Flag),
						Type:       col.Tp.CompactStr(),
						DataType:   types.TypeToStr(col.Tp.Tp, col.Tp.Charset),
						Length:     length,
					})
				}

				out[tableName] = table
			}
		}

		return out, nil
	}
}
