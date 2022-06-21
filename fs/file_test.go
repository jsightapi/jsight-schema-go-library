package fs

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestNewFile(t *testing.T) {
	const name = "foo"
	var content = bytes.Bytes("content")

	f := NewFile(name, content)

	assert.Equal(t, name, f.name)
	assert.Equal(t, name, f.Name())
	assert.Equal(t, content, f.content)
	assert.Equal(t, content, f.Content())
}
