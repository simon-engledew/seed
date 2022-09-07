MySQL load test data generator. Automatically generates suitable data by parsing schemas.

Example:

```go
package main

import (
	"database/sql"
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/distribution"
	"github.com/simon-engledew/seed/inspectors"
	"os"
	"context"
)

func generate(db *sql.DB) {
	schema, err := seed.Build(inspectors.InspectMySQLConnection(db))
	if err != nil {
		panic(err)
	}
	generator := schema.Generator(context.Background(), consumers.MySQLInsertWriter(os.Stdout, 100))
	// generate between 100 and 300 owners
	generator.Insert("owners", distribution.Range(100, 200), func(g *seed.RowGenerator) {
		// generate cats for 3/10 owners
		g.Insert("cats", distribution.Ratio(0.3))
	})
	if err := generator.Wait(); err != nil {
		panic(err)
	}
}
```
