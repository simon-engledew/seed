package seed_test

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/consumers"
	"github.com/simon-engledew/seed/distribution"
	"github.com/simon-engledew/seed/generators"
	"github.com/simon-engledew/seed/inspectors"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"strconv"
	"testing"
)

func TestBuild(t *testing.T) {
	testInspector := func(fn func(tableName, columnName string, column inspectors.ColumnInfo)) error {
		fn("test", "id", inspectors.MySQLColumn{
			Name:       "id",
			DataType:   "bigint",
			IsPrimary:  true,
			IsUnsigned: true,
			Length:     20,
			ColumnType: "bigint(20)",
		})
		return nil
	}
	s, err := seed.Build(testInspector)
	require.NoError(t, err)
	s.Transform(
		seed.ReplaceColumnType("tinyint(1)", generators.Faker(func(faker *gofakeit.Faker) (string, bool) {
			return strconv.FormatBool(faker.Bool()), false
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
