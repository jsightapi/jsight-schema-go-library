package constraint

import (
	jschema "j/schema"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequiredKeys_Type(t *testing.T) {
	assert.Equal(t, RequiredKeysConstraintType, NewRequiredKeys().Type())
}

func TestRequiredKeys_ASTNode(t *testing.T) {
	c := NewRequiredKeys()
	c.AddKey("foo")
	c.AddKey("bar")

	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeArray,
		Properties: &jschema.RuleASTNodes{},
		Items: []jschema.RuleASTNode{
			{
				JSONType:   jschema.JSONTypeString,
				Value:      "foo",
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			},
			{
				JSONType:   jschema.JSONTypeString,
				Value:      "bar",
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			},
		},
		Source: jschema.RuleASTNodeSourceManual,
	}, c.ASTNode())
}
