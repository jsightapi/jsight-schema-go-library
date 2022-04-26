package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseNode_SetRealType(t *testing.T) {
	n := &baseNode{}

	n.SetRealType("foo")
	assert.Equal(t, "foo", n.realType)
}

func TestBaseNode_RealType(t *testing.T) {
	n := &baseNode{
		realType: "foo",
	}

	assert.Equal(t, "foo", n.RealType())
}
