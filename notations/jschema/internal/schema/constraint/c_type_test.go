package constraint

import (
	jschema "j/schema"
	"j/schema/bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypeConstraint_Type(t *testing.T) {
	assert.Equal(t,
		TypeConstraintType,
		NewType(bytes.Bytes("foo"), jschema.RuleASTNodeSourceGenerated).Type(),
	)
}

func TestTypeConstraint_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeString,
		Value:      "foo",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceGenerated,
	}, NewType(bytes.Bytes("foo"), jschema.RuleASTNodeSourceGenerated).ASTNode())
}
