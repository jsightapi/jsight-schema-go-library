package bytes

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBlank(t *testing.T) {
	testByteIsserFunction(t, IsBlank, ' ', '\t', '\n', '\r')
}

func TestIsSpace(t *testing.T) {
	testByteIsserFunction(t, IsSpace, ' ', '\t')
}

func TestIsNewLine(t *testing.T) {
	testByteIsserFunction(t, IsNewLine, '\n', '\r')
}

func TestIsDigit(t *testing.T) {
	testByteIsserFunction(t, IsDigit, '0', '1', '2', '3', '4', '5', '6', '7', '8', '9')
}

func TestIsHexDigit(t *testing.T) {
	testByteIsserFunction(t, IsHexDigit,
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
		'a', 'b', 'c', 'd', 'e', 'f',
		'A', 'B', 'C', 'D', 'E', 'F',
	)
}

func TestIsValidUserTypeNameByte(t *testing.T) {
	testByteIsserFunction(t, IsValidUserTypeNameByte,
		// Allowed special symbols.
		'-', '_',

		// Any english lower letters.
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o',
		'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',

		// Any english capital letters.
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O',
		'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',

		// Any digits.
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	)
}

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

func testByteIsserFunction(t *testing.T, tested func(byte) bool, valid ...byte) {
	t.Helper()

	for c := byte(0); c < 255; c++ {
		t.Run(string(c), func(t *testing.T) {
			actual := tested(c)
			assert.Equal(t, bytes.Contains(valid, []byte{c}), actual)
		})
	}
}
