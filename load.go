package seed

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/jinzhu/inflection"
	"github.com/jpillora/longestcommon"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	_ "github.com/pingcap/tidb/types/parser_driver"
	"github.com/simon-engledew/seed/generators"
)

type Column struct {
	Name      string
	Generator generators.ColumnGenerator
}

// Load reads in MySQL schemas on stdin, returning a test data generator.
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

				table = append(table, &Column{
					Name:      columnName,
					Generator: generators.Column(col.Tp, isPrimary),
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
							columns[i].Generator = generators.Reference(tableName, j)
						}
					}
				}
			}
		}
	}

	return schema, nil
}
