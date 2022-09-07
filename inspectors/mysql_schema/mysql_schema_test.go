package mysql_schema_test

import (
	"bytes"
	"context"
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/distribution"
	"github.com/stretchr/testify/require"
	"io"
	"mysql_schema"
	"os"
	"strings"
	"testing"
)

func TestExample(t *testing.T) {
	i := mysql_schema.InspectMySQLSchema(strings.NewReader(`
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
	schema, err := seed.Build(i)
	require.NoError(t, err)
	var buf bytes.Buffer
	generator := schema.Generator(context.Background(), consumers.MySQLInsertWriter(io.MultiWriter(os.Stdout, &buf), 100))
	// generate between 100 and 300 owners
	generator.Insert("owners", distribution.Range(100, 200), func(g *seed.RowGenerator) {
		// generate cats for 3/10 owners
		g.Insert("cats", distribution.Ratio(0.3))
	})
	require.NoError(t, generator.Wait())
	require.NotEmpty(t, buf.Bytes())
}
