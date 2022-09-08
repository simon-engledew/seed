package seed

import (
	"github.com/jinzhu/inflection"
	"github.com/jpillora/longestcommon"
	"strings"
)

func Build(schema map[string][]Column) (Schema, error) {
	tableNames := make([]string, 0, len(schema))

	for tableName := range schema {
		tableNames = append(tableNames, tableName)
	}

	prefix := longestcommon.Prefix(tableNames)

	for _, columns := range schema {
		for i, column := range columns {
			columnName := column.Name()
			if strings.HasSuffix(columnName, "_id") {
				tableName := prefix + inflection.Plural(columnName[:len(columnName)-3])

				if parent, ok := schema[tableName]; ok {
					for j, target := range parent {
						if target.Name() == "id" {
							columns[i] = overrideColumn(columns[i], Reference(tableName, j))
						}
					}
				}
			}
		}
	}

	return schema, nil
}
