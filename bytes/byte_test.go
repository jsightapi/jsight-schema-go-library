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

func testByteIsserFunction(t *testing.T, tested func(byte) bool, valid ...byte) {
	t.Helper()

	for c := byte(0); c < 255; c++ {
		t.Run(string(c), func(t *testing.T) {
			actual := tested(c)
			assert.Equal(t, bytes.Contains(valid, []byte{c}), actual)
		})
	}
}
