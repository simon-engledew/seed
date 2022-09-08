package distribution_test

import (
	"testing"

	"github.com/shaaraddalvi/seed/distribution"
	"github.com/stretchr/testify/require"
)

func TestRatioZero(t *testing.T) {
	fn := distribution.Ratio(0)
	require.False(t, fn())
}

func TestRatioOne(t *testing.T) {
	fn := distribution.Ratio(1)
	require.True(t, fn())
}

func TestFixed(t *testing.T) {
	fn := distribution.Fixed(3)
	for i := 0; i < 4; i++ {
		require.Equal(t, i < 3, fn())
	}
}

func TestRangeOne(t *testing.T) {
	fn := distribution.Range(1, 1)
	require.True(t, fn())
	require.False(t, fn())
}

func TestRangeZero(t *testing.T) {
	fn := distribution.Range(0, 0)
	require.False(t, fn())
}
