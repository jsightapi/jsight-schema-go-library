package scanner

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestQuoteChar(t *testing.T) {
	cc := map[byte]string{
		'\'': "'\\''",
		'"':  `'"'`,
		'c':  "'c'",
	}

	for given, expected := range cc {
		t.Run(string(given), func(t *testing.T) {
			actual := QuoteChar(given)
			assert.Equal(t, expected, actual)
		})
	}
}
