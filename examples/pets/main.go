package main

import (
	"context"
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/distribution"
	"github.com/simon-engledew/seed/generators"
	"github.com/simon-engledew/seed/inspectors/schema/mysql"
	"os"
	"strings"
)

func main() {
	def, err := mysql.InspectMySQLSchema(strings.NewReader(`
	CREATE TABLE owners (
		id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
		name VARCHAR(255),
		PRIMARY KEY (id)
	);
    	CREATE TABLE cats (
		id BIGINT UNSIGNED,
		name VARCHAR(255),
		owner_id BIGINT,
		name VARCHAR(255),
		PRIMARY KEY (id)
    	);
	`))
	if err != nil {
		panic(err)
	}
	schema, err := seed.Build(def)
	if err != nil {
		panic(err)
	}
	schema.Transform(seed.Merge(map[string]map[string]consumers.ValueGenerator{
		"cats": {
			"name": generators.Format[generators.Quoted]("{beername}"),
		},
	}))
	generator := schema.Generator(context.Background(), consumers.MySQLInsertWriter(os.Stdout, 100))
	// generate between 100 and 200 owners
	generator.Insert("owners", distribution.Range(100, 200), func(g *seed.RowGenerator) {
		// generate cats for 3/10 owners
		g.Insert("cats", distribution.Ratio(0.3))
	})
	if err := generator.Wait(); err != nil {
		panic(err)
	}
}
