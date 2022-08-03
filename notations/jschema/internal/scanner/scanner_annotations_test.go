package scanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScanner_isAnnotationStart(t *testing.T) {
	cc := map[byte]bool{
		'/': true,
	}
	for i := 0; i <= 255; i++ {
		if i != '/' {
			cc[byte(i)] = false
		}
	}

	s := &Scanner{}
	for c, expected := range cc {
		t.Run(string(c), func(t *testing.T) {
			assert.Equal(t, expected, s.isAnnotationStart(c))
		})
	}
}
