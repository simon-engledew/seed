package escape_test

import (
	"github.com/simon-engledew/seed/escape"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestQuote(t *testing.T) {
	require.Equal(t, "'hel\\'lo'", escape.Quote("hel'lo"))
}

func TestQuoteIdentifier(t *testing.T) {
	require.Equal(t, "`hel``lo`", escape.QuoteIdentifier("hel`lo"))
}
