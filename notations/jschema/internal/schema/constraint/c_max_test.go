package constraint

import (
	"github.com/stretchr/testify/assert"
	"j/schema"
	"j/schema/bytes"
	"testing"
)

func TestMax_Type(t *testing.T) {
	assert.Equal(t, MaxConstraintType, NewMax(bytes.Bytes("1")).Type())
}

func TestMax_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewMax(bytes.Bytes("1")).ASTNode())
}
