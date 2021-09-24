package seed_test

import (
	"github.com/simon-engledew/seed"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	s, err := seed.Load(strings.NewReader("CREATE TABLE test (id BIGINT(20) NOT NULL AUTO_INCREMENT)"))
	require.NoError(t, err)
	s.Transform(func(table string, c *seed.Column) {
		require.Equal(t, "test", table)
		require.Equal(t, "id", c.Name)
	})
}
