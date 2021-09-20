package main

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/distribution"
	"github.com/simon-engledew/seed/generators"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func panicOnErr(fn func() error) {
	if err := fn(); err != nil {
		panic(err)
	}
}

func ReplaceColumns(columns map[string]generators.ValueGenerator) seed.SchemaTransform {
	return func(t string, c *seed.Column) {
		if generator, ok := columns[c.Name]; ok {
			c.Generator = generator
		}
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

	schema.Transform(ReplaceColumns(map[string]generators.ValueGenerator{
		"repository_id":    repoID,
		"user_id":          generators.Identity(gofakeit.Generate("{number:1,256}")),
		"version":          generators.Format("#.#.#"),
		"semantic_version": generators.Format("#.#.#"),
		"guid":             generators.Format("{uuid}"),
		"canonical_name":   generators.Format("{hackernoun}"),
	}))

	os.MkdirAll("out", 0o755)

	generator := schema.Generator(func(t string, c []string, rows chan []string) {
		go func() {
			f, err := os.Create(filepath.Join("out", t+".sql"))
			if err != nil {
				panic(err)
			}
			defer panicOnErr(f.Close)
			row := <-rows
			fmt.Fprintf(f, "INSERT INTO %s (%s) VALUES (%s)", t, strings.Join(c, ", "), strings.Join(row, ", "))
			for row := range rows {
				fmt.Fprintf(f, ", (%s)", strings.Join(row, ", "))
			}
			fmt.Fprintf(f, ";")
		}()
	})

	for ; repositoryID < 1000; repositoryID += 1 {
		once := distribution.Fixed(1)

		generator.Insert("ts_tools", distribution.Random(1, 3), func(g seed.Generator) {
			g.Insert("ts_tool_versions", distribution.Fixed(3), func(g seed.Generator) {
				g.Insert("ts_analyses", once, func(g seed.Generator) {
					fmt.Println("ts_analyses x 1")
					g.Insert("ts_logical_alerts_seq", distribution.Fixed(1))
					//fmt.Println("ts_rules x 100")

					g.Insert("ts_rules", distribution.Fixed(100), func(g seed.Generator) {
						g.Insert("ts_rule_tags", distribution.Ratio(0.8))

						g.Insert("ts_analysis_rules", distribution.Fixed(1))
						g.Insert("ts_snippets", distribution.Fixed(1), func(g seed.Generator) {
							//fmt.Println("ts_logical_alerts x 1000")

							g.Insert("ts_logical_alerts", distribution.Fixed(1000), func(g seed.Generator) {
								g.Insert("ts_timeline_events", distribution.Random(0, 1))
								g.Insert("ts_physical_alerts", distribution.Fixed(1))
							})
						})
					})
				})
			})
		})
	}
}
