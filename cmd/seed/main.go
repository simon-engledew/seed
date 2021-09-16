package main

import (
	"context"
	"fmt"
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/distribution"
	"github.com/simon-engledew/seed/generators"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type key string

var repoKey key = "repoID"

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

func Merge(schema map[string]map[string]generators.ValueGenerator) seed.SchemaTransform {
	return func(t string, c *seed.Column) {
		if i, ok := schema[t]; ok {
			if j, ok := i[c.Name]; ok {
				c.Generator = j
			}
		}
	}
}

func main() {
	schema, err := seed.Load(os.Stdin)
	if err != nil {
		panic(err)
	}

	repositoryID := uint64(0)

	schema.Transform(ReplaceColumns(map[string]generators.ValueGenerator{
		"repository_id": generators.Func(func(ctx context.Context) string {
			return strconv.FormatUint(ctx.Value(repoKey).(uint64), 10)
		}),
		"user_id":          generators.Format("{number:1,256}"),
		"version":          generators.Format("#.#.#"),
		"semantic_version": generators.Format("#.#.#"),
		"guid":             generators.Format("{uuid}"),
		"canonical_name":   generators.Format("{hackernoun}"),
	}))

	err = os.MkdirAll("out", 0o755)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	var wg sync.WaitGroup

	generator := schema.Generator(ctx, func(t string, c []string, rows chan []string) {
		wg.Add(1)
		go func() {
			counter := 0

			f, err := os.Create(filepath.Join("out", t+".sql"))
			if err != nil {
				panic(err)
			}
			defer panicOnErr(f.Close)
			defer wg.Done()

			row := <-rows
			if len(row) > 0 {
				fmt.Fprintf(f, "INSERT INTO %s (%s) VALUES (%s)", t, strings.Join(c, ", "), strings.Join(row, ", "))
				counter += 1

				for row := range rows {
					fmt.Fprintf(f, ",\n(%s)", strings.Join(row, ", "))

					counter += 1
				}
				fmt.Fprintf(f, ";")

				fmt.Printf("%s x %d\n", t, counter)
			}
		}()
	})

	for ; repositoryID < 100; repositoryID += 1 {
		once := distribution.Fixed(1)
		alerts := distribution.Fixed(100)

		ctx := context.WithValue(ctx, repoKey, repositoryID)

		generator.InsertContext(ctx, "ts_tools", distribution.Range(1, 3), func(g seed.Generator) {
			g.Insert("ts_tool_versions", distribution.Fixed(3), func(g seed.Generator) {
				g.Insert("ts_analyses", once, func(g seed.Generator) {

					g.Insert("ts_logical_alerts_seq", distribution.Fixed(1))

					g.Insert("ts_rules", distribution.Fixed(100), func(g seed.Generator) {
						g.Insert("ts_rule_tags", distribution.Ratio(0.8))

						g.Insert("ts_analysis_rules", distribution.Fixed(1))
						g.Insert("ts_snippets", distribution.Fixed(1), func(g seed.Generator) {

							g.Insert("ts_logical_alerts", alerts, func(g seed.Generator) {
								g.Insert("ts_timeline_events", distribution.Ratio(0.1))
								g.Insert("ts_physical_alerts", distribution.Range(8, 10))
							})
						})
					})
				})
			})
		})
	}

	generator.Done()
	wg.Wait()
}
