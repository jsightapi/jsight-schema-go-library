package constraint

import (
	"j/schema"
	"j/schema/bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMin_Type(t *testing.T) {
	assert.Equal(t, MinConstraintType, NewMin(bytes.Bytes("1")).Type())
}

func TestMin_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewMin(bytes.Bytes("1")).ASTNode())
}
