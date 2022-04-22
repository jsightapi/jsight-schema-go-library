package constraint

import (
	"j/schema"
	"j/schema/bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinLength_Type(t *testing.T) {
	assert.Equal(t, MinLengthConstraintType, NewMinLength(bytes.Bytes("1")).Type())
}

func TestMinLength_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewMinLength(bytes.Bytes("1")).ASTNode())
}
