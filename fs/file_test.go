package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestNewFile(t *testing.T) {
	var expected = bytes.Bytes("content")

	t.Run("string", func(t *testing.T) {
		testNewFile(t, "content", expected)
	})

	t.Run("[]byte", func(t *testing.T) {
		testNewFile(t, []byte("content"), expected)
	})

	t.Run("bytes.Bytes", func(t *testing.T) {
		testNewFile(t, bytes.Bytes("content"), expected)
	})
}

func testNewFile[T FileContent](t *testing.T, given T, expected bytes.Bytes) {
	const name = "foo"

	f, err := NewFile(name, given)
	require.NoError(t, err)

	assert.Equal(t, name, f.name)
	assert.Equal(t, name, f.Name())
	assert.Equal(t, expected, f.content)
	assert.Equal(t, expected, f.Content())
}

func Test_normalizeFileContent(t *testing.T) {
	var expected = bytes.Bytes("content")

	t.Run("string", func(t *testing.T) {
		testNormalizeFileContent(t, "content", expected)
	})

	t.Run("[]byte", func(t *testing.T) {
		testNormalizeFileContent(t, []byte("content"), expected)
	})

	t.Run("bytes.Bytes", func(t *testing.T) {
		testNormalizeFileContent(t, bytes.Bytes("content"), expected)
	})

	t.Run("nil []byte", func(t *testing.T) {
		var b []byte
		testNormalizeFileContent(t, b, nil)
	})

	t.Run("nil bytes.Bytes", func(t *testing.T) {
		var b bytes.Bytes
		testNormalizeFileContent(t, b, nil)
	})
}

func testNormalizeFileContent[T FileContent](t *testing.T, given T, expected bytes.Bytes) {
	actual := normalizeFileContent(given)

	assert.Equal(t, expected, actual)
}
