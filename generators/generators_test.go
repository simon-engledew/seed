package generators_test

import (
	"context"
	"github.com/simon-engledew/seed/generators"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestText(t *testing.T) {
	v := generators.Format("test-{number:1,1}").Value(context.Background())
	require.Equal(t, "test-1", v.String())
	require.True(t, v.Escape())
}
