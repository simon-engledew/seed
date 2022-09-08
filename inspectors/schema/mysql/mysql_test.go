package mysql_test

import (
	"bytes"
	"context"
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/distribution"
	"github.com/simon-engledew/seed/inspectors/schema/mysql"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"strings"
	"testing"
)

func TestExample(t *testing.T) {
	tables, err := mysql.InspectMySQLSchema(strings.NewReader(`
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
	require.NoError(t, err)
	schema, err := seed.Build(tables)
	require.NoError(t, err)
	var buf bytes.Buffer
	generator := schema.Generator(context.Background(), consumers.MySQLInsertWriter(io.MultiWriter(os.Stdout, &buf), 100))
	// generate between 100 and 300 owners
	generator.Insert("owners", distribution.Range(10, 20), func(g *seed.RowGenerator) {
		// generate cats for 3/10 owners
		g.Insert("cats", distribution.Ratio(0.3))
	})
	require.NoError(t, generator.Wait())
	require.NotEmpty(t, buf.Bytes())
}
