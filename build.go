package seed

import (
	"strings"

	"github.com/jinzhu/inflection"
	"github.com/jpillora/longestcommon"
	"github.com/shaaraddalvi/seed/inspectors"
)

func Build(i inspectors.Inspector) (Schema, error) {
	schema := make(map[string][]*Column)

	err := i(func(tableName, columnName string, columnInfo inspectors.ColumnInfo) {
		if _, ok := schema[tableName]; !ok {
			schema[tableName] = make([]*Column, 0)
		}
		schema[tableName] = append(schema[tableName], &Column{
			Name:      columnName,
			Generator: columnInfo.Generator(),
			Type:      columnInfo.Type(),
		})
	})
	if err != nil {
		return nil, err
	}

	tableNames := make([]string, 0, len(schema))

	for tableName := range schema {
		tableNames = append(tableNames, tableName)
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
