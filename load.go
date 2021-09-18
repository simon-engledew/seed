package seed

import (
	"fmt"
	"github.com/jinzhu/inflection"
	"github.com/jpillora/longestcommon"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"github.com/simon-engledew/seed/generators"
	"io"
	"io/ioutil"
	"strings"
)

type TableName string
type ColumnName string
type Row map[ColumnName]string
type Rows map[TableName]Row

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

				table[columnName] = generators.Get(col.Tp, isPrimary)
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
