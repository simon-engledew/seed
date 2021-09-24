package generators_test

import (
	"context"
	"github.com/simon-engledew/seed/generators"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestText(t *testing.T) {
	require.Equal(t, "'test-1'", generators.Format("test-{number:1,1}").Value(context.Background()))
}
