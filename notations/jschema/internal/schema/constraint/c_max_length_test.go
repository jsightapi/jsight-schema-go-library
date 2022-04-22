package constraint

import (
	"j/schema"
	"j/schema/bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxLength_Type(t *testing.T) {
	assert.Equal(t, MaxLengthConstraintType, NewMaxLength(bytes.Bytes("1")).Type())
}

func TestMaxLength_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewMaxLength(bytes.Bytes("1")).ASTNode())
}
