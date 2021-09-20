package main

import (
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/distribution"
	"github.com/simon-engledew/seed/generators"
	"os"
	"strconv"
)

func ReplaceColumns(columns map[seed.ColumnName]generators.ValueGenerator) seed.SchemaTransform {
	return func(t seed.TableName, c seed.ColumnName, g generators.ValueGenerator) generators.ValueGenerator {
		if generator, ok := columns[c]; ok {
			return generator
		}
		return g
	}
}

//func Merge(schema seed.Schema) seed.SchemaTransform {
//	return func(t seed.TableName, c seed.ColumnName, g generators.ValueGenerator) generators.ValueGenerator {
//		if i, ok := schema[t]; ok {
//			if j, ok := i[c]; ok {
//				return j
//			}
//		}
//		return g
//	}
//}

func main() {
	schema, err := seed.Load(os.Stdin)
	if err != nil {
		panic(err)
	}

	repositoryID := 0

	repoID := generators.Func(func() string {
		return strconv.Itoa(repositoryID)
	})

	schema.Transform(ReplaceColumns(map[seed.ColumnName]generators.ValueGenerator{
		"repository_id":    repoID,
		"user_id":          generators.Identity(gofakeit.Generate("{number:1,256}")),
		"version":          generators.Format("#.#.#"),
		"semantic_version": generators.Format("#.#.#"),
		"guid":             generators.Format("{uuid}"),
		"canonical_name":   generators.Format("{hackernoun}"),
	}))

	for ; repositoryID < 1000; repositoryID += 1 {
		once := distribution.Fixed(1)

		generator := schema.Generator(os.Stdout)
		generator.Insert("ts_tools", distribution.Random(1, 3), func(tools seed.Generator) {
			tools.Insert("ts_tool_versions", distribution.Fixed(3), func(toolVersions seed.Generator) {
				toolVersions.Insert("ts_analyses", once, func(analyses seed.Generator) {
					analyses.Insert("ts_timeline_events", distribution.Random(1, 10))
					analyses.Insert("ts_logical_alerts_seq", distribution.Fixed(1))
					analyses.Insert("ts_rules", distribution.Fixed(800), func(rules seed.Generator) {
						rules.Insert("ts_rule_tags", distribution.Ratio(0.8))

						rules.Insert("ts_analysis_rules", distribution.Fixed(1))

						rules.Insert("ts_logical_alerts", distribution.Fixed(125), func(physicalAlerts seed.Generator) {
							physicalAlerts.Insert("ts_physical_alerts", distribution.Fixed(1))
						})
					})
				})
			})
		})
	}
}
