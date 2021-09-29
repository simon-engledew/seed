package seed_test

import (
	"context"
	"github.com/simon-engledew/seed"
	"github.com/simon-engledew/seed/distribution"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	s, err := seed.Load(strings.NewReader("CREATE TABLE test (id BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY)"))
	require.NoError(t, err)
	s.Transform(func(table string, c *seed.Column) {
		require.Equal(t, "test", table)
		require.Equal(t, "id", c.Name)
		require.Equal(t, "bigint(20)", c.Type)
	})
	testConsumer := func(g *errgroup.Group) func(t string, c []string, rows chan []string) {
		return func(table string, c []string, rows chan []string) {
			g.Go(func() error {
				require.Equal(t, "test", table)
				require.ElementsMatch(t, []string{"`id`"}, c)
				row := <-rows
				require.ElementsMatch(t, []string{"1"}, row)
				return nil
			})
		}
	}

	g := s.Generator(context.Background(), testConsumer)
	g.Insert("test", distribution.Fixed(1))
	require.NoError(t, g.Wait())
}
