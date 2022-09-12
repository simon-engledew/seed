package seed_test

import (
	"context"
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/distribution"
	"github.com/simon-engledew/seed/generators"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestUnique(t *testing.T) {
	schema := make(seed.Schema)
	schema["test"] = []*seed.Column{
		{Name: "a", Type: "bigint", Generator: generators.UniqueRow(generators.Format[generators.Unquoted]("{number:1,}"), "b")},
		{Name: "b", Type: "bigint", Generator: generators.Unique(generators.Format[generators.Unquoted]("{number:1,}"))},
	}
	generator := schema.Generator(context.Background(), consumers.MySQLInsertWriter(os.Stdout, 100))
	generator.Insert("test", distribution.Fixed(10))
	require.NoError(t, generator.Wait())
}

func TestLazy(t *testing.T) {
	fn := func(_ context.Context, row map[string]consumers.Value) consumers.Value {
		return row["a"].(generators.UnsignedInt) + 2
	}

	schema := make(seed.Schema)
	schema["test"] = []*seed.Column{
		{Name: "a", Type: "bigint", Generator: generators.Column("bigint", true, 20)},
		{Name: "b", Type: "bigint", Generator: generators.Lazy(fn)},
	}
	generator := schema.Generator(context.Background(), consumers.MySQLInsertWriter(os.Stdout, 100))
	generator.Insert("test", distribution.Fixed(10))
	require.NoError(t, generator.Wait())
}
