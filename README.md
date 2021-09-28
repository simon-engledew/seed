MySQL load test data generator. Automatically generates suitable data by parsing schemas.

Example:

```go
package main

import (
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/distribution"
	"os"
	"context"
)

func main() {
	schema, err := seed.Load(os.Stdin)
	if err != nil {
		panic(err)
	}
	generator := schema.Generator(context.Background(), consumers.Inserts(os.Stdout, 100))
	generator.Insert("owners", distribution.Range(100, 200), func(g *seed.Generator) {
		g.Insert("cats", distribution.Ratio(0.3))
	})
	err := generator.Wait()
	if err != nil {
		panic(err)
	}
}
```
