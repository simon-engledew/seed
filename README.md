MySQL load test data generator. Automatically generates suitable data by parsing schemas.

Example:

```go
package main

import (
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/distribution"
	"github.com/simon-engledew/seed/inspectors"
	"os"
	"context"
)

func main() {
	schema, err := seed.Build(inspectors.InspectMySQLSchema(os.Stdin))
	if err != nil {
		panic(err)
	}
	generator := schema.Generator(context.Background(), consumers.InsertWriter(os.Stdout, 100))
	generator.Insert("owners", distribution.Range(100, 200), func(g *seed.RowGenerator) {
		g.Insert("cats", distribution.Ratio(0.3))
	})
	if err := generator.Wait(); err != nil {
		panic(err)
	}
}
```
