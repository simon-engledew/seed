package seed_test

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/distribution"
	"github.com/simon-engledew/seed/generators"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"testing"
)

func TestBuild(t *testing.T) {
	columns := []byte(`{"test":[{"name":"id","data_type":"bigint","is_primary":true,"is_unsigned":true,"length":20,"column_type":"bigint(20)"}]}`)
	s, err := seed.Build(columns)
	require.NoError(t, err)
	s.Transform(
		seed.ReplaceColumnType("tinyint(1)", generators.Faker(func(faker *gofakeit.Faker) consumers.Value {
			return generators.Bool(faker.Bool())
		})),
	)
	s.Transform(func(table string, c *seed.Column) {
		require.Equal(t, "test", table)
		require.Equal(t, "id", c.Name)
		require.Equal(t, "bigint(20)", c.Type)
	})
	testConsumer := consumers.MySQLConsumer(func(ctx context.Context, g *errgroup.Group) func(t string, c []string, rows chan []string) {
		return func(table string, c []string, rows chan []string) {
			g.Go(func() error {
				require.Equal(t, "`test`", table)
				require.ElementsMatch(t, []string{"`id`"}, c)
				row := <-rows
				require.ElementsMatch(t, []string{"1"}, row)
				return nil
			})
		}
	})

	g := s.Generator(context.Background(), testConsumer)
	g.Insert("test", distribution.Fixed(1))
	require.NoError(t, g.Wait())
}
