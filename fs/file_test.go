package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestNewFile(t *testing.T) {
	var expected = bytes.NewBytes("content")

	t.Run("string", func(t *testing.T) {
		testNewFile(t, "content", expected)
	})

	t.Run("[]byte", func(t *testing.T) {
		testNewFile(t, []byte("content"), expected)
	})

	t.Run("bytes.Bytes", func(t *testing.T) {
		testNewFile(t, bytes.NewBytes("content"), expected)
	})
}

func testNewFile[T bytes.Byter](t *testing.T, given T, expected bytes.Bytes) {
	const name = "foo"

	f := NewFile(name, given)

	assert.Equal(t, name, f.name)
	assert.Equal(t, name, f.Name())
	assert.Equal(t, expected, f.content)
	assert.Equal(t, expected, f.Content())
}
