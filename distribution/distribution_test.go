package distribution_test

import (
	"github.com/simon-engledew/seed/distribution"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRangeOne(t *testing.T) {
	fn := distribution.Range(1, 1)
	require.True(t, fn())
	require.False(t, fn())
}

func TestRangeZero(t *testing.T) {
	fn := distribution.Range(0, 0)
	require.False(t, fn())
}
