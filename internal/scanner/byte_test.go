package scanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func BenchmarkQuoteChar(b *testing.B) {
	cc := []byte{
		'\'',
		'"',
		'c',
	}

	b.ReportAllocs()

	for _, c := range cc {
		b.Run(string(c), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				QuoteChar(c)
			}
		})
	}
}
