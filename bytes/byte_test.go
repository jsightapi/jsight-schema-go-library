package bytes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBlank(t *testing.T) {
	for c := byte(0); c < 255; c++ {
		t.Run(string(c), func(t *testing.T) {
			actual := IsBlank(c)
			assert.Equal(t, c == ' ' || c == '\t' || c == '\n' || c == '\r', actual)
		})
	}
}

func TestIsSpace(t *testing.T) {
	for c := byte(0); c < 255; c++ {
		t.Run(string(c), func(t *testing.T) {
			actual := IsSpace(c)
			assert.Equal(t, c == ' ' || c == '\t', actual)
		})
	}
}

func TestIsNewLine(t *testing.T) {
	for c := byte(0); c < 255; c++ {
		t.Run(string(c), func(t *testing.T) {
			actual := IsNewLine(c)
			assert.Equal(t, c == '\n' || c == '\r', actual)
		})
	}
}
