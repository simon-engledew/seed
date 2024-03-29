MySQL load test data generator. Automatically generates suitable data by parsing schemas.

Example:

```go
package main

import (
	"context"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/distribution"
	"github.com/simon-engledew/seed/inspectors/db/mysql"
	"os"
)

func main() {
	db, err := sql.Open("mysql", os.Args[1])
	if err != nil {
		panic(err)
	}

	def, err := mysql.InspectMySQLConnection(db)
	if err != nil {
		panic(err)
	}
	schema, err := seed.Build(def)
	if err != nil {
		panic(err)
	}
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
```

There is also an inspector that can parse mysql schema files and generate data without a database connection.

```go
package main

import (
	"context"
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/distribution"
	"github.com/simon-engledew/seed/inspectors/schema/mysql"
	"os"
	"strings"
)

func main() {
	def, err := mysql.InspectMySQLSchema(strings.NewReader(`
	CREATE TABLE owners (
		id BIGINT UNSIGNED,
		name VARCHAR(255)
	);

    	CREATE TABLE cats (
		id BIGINT UNSIGNED,
		owner_id BIGINT,
		name VARCHAR(255)
	);
	`))
	if err != nil {
		panic(err)
	}
	schema, err := seed.Build(def)
	if err != nil {
		panic(err)
	}
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
```
