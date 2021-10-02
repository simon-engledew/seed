package seed

import (
	"github.com/jinzhu/inflection"
	"github.com/jpillora/longestcommon"
	"github.com/simon-engledew/seed/inspectors"
	"strings"
)

func Build(i inspectors.Inspector) (Schema, error) {
	info, err := i()
	if err != nil {
		return nil, err
	}

	schema := make(map[string][]*Column)

	tableNames := make([]string, 0, len(info))

	for tableName, columns := range info {
		table := make([]*Column, 0, len(columns))
		tableNames = append(tableNames, tableName)

		for columnName, column := range columns {
			table = append(table, &Column{
				Name:      columnName,
				Generator: column.Generator(),
				Type:      column.Type(),
			})
		}

		schema[tableName] = table
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
