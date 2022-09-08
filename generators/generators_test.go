package generators_test

import (
	"context"
	"testing"

	"github.com/shaaraddalvi/seed/generators"
	"github.com/stretchr/testify/require"
)

func TestText(t *testing.T) {
	v := generators.Format("test-{number:1,1}").Value(context.Background())
	require.Equal(t, "test-1", v.Value)
	require.True(t, v.Quote)
}
