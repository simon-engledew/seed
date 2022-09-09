package seed

import (
	"encoding/json"
	"github.com/jinzhu/inflection"
	"github.com/jpillora/longestcommon"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/generators"
	"golang.org/x/exp/maps"
	"strings"
)

type ColumnDefinition struct {
	Name       string `json:"name"`
	DataType   string `json:"data_type"`
	IsPrimary  bool   `json:"is_primary"`
	IsUnsigned bool   `json:"is_unsigned"`
	Length     int    `json:"length"`
	Type       string `json:"column_type"`
}

func (c ColumnDefinition) Generator() consumers.ValueGenerator {
	if c.IsPrimary {
		return generators.Counter()
	}
	if gen := generators.Column(c.DataType, c.IsUnsigned, c.Length); gen != nil {
		return gen
	}
	return generators.Identity[generators.Unquoted]("NULL")
}

func Build(data []byte) (Schema, error) {
	var tables map[string][]ColumnDefinition

	if err := json.Unmarshal(data, &tables); err != nil {
		return nil, err
	}

	schema := make(Schema, len(tables))

	for tableName, columns := range tables {
		schema[tableName] = make([]*Column, 0, len(columns))
		for _, c := range columns {
			schema[tableName] = append(schema[tableName], &Column{
				Name:      c.Name,
				Type:      c.Type,
				Generator: c.Generator(),
			})
		}
	}

	prefix := longestcommon.Prefix(maps.Keys(tables))

	for _, columns := range schema {
		for i, column := range columns {
			columnName := column.Name
			if strings.HasSuffix(columnName, "_id") {
				tableName := prefix + inflection.Plural(columnName[:len(columnName)-3])

				if parent, ok := tables[tableName]; ok {
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
