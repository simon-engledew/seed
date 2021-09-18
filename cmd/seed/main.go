package main

import (
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/distribution"
	"github.com/simon-engledew/seed/generators"
	"os"
)

func ReplaceColumn(columnName seed.ColumnName, generator generators.ValueGenerator) seed.SchemaTransform {
	return func(t seed.TableName, c seed.ColumnName, g generators.ValueGenerator) generators.ValueGenerator {
		if c == columnName {
			return generator
		}
		return g
	}
}

func main() {
	schema, err := seed.Load(os.Stdin)
	if err != nil {
		panic(err)
	}

	schema.Transform(ReplaceColumn("repository_id", generators.Identity("1")))
	schema.Merge(seed.Schema{
		"ts_tool_versions": {
			"version": generators.TextFormat("#.#.#"),
		},
	})

	once := distribution.Fixed(1)

	generator := schema.Generator(os.Stdout)
	generator.Insert("ts_tools", distribution.Fixed(3), func(tools seed.Generator) {
		tools.Insert("ts_tool_versions", distribution.Fixed(3), func(toolVersions seed.Generator) {
			toolVersions.Insert("ts_analyses", once, func(analyses seed.Generator) {
				analyses.Insert("ts_logical_alerts_seq", distribution.Fixed(1))
				analyses.Insert("ts_rules", distribution.Fixed(5), func(rules seed.Generator) {
					rules.Insert("ts_rule_tags", distribution.Ratio(0.8))

					rules.Insert("ts_analysis_rules", distribution.Fixed(1))

					rules.Insert("ts_logical_alerts", distribution.Ratio(0.33), func(physicalAlerts seed.Generator) {
						physicalAlerts.Insert("ts_physical_alerts", distribution.Fixed(1))
					})
				})
			})
		})
	})
}
