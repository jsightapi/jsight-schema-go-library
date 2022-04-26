package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
)

func TestTypesList_Type(t *testing.T) {
	assert.Equal(t, TypesListConstraintType, NewTypesList(jschema.RuleASTNodeSourceManual).Type())
}

func TestTypesList_ASTNode(t *testing.T) {
	l := NewTypesList(jschema.RuleASTNodeSourceManual)

	an := jschema.RuleASTNode{
		JSONType: jschema.JSONTypeObject,
		Properties: jschema.NewRuleASTNodes(
			map[string]jschema.RuleASTNode{
				"type": newRuleASTNode(jschema.JSONTypeString, "foo", jschema.RuleASTNodeSourceManual),
			},
			[]string{"type"},
		),
	}

	l.AddNameWithASTNode("foo", "foo", an)
	l.AddName("bar", "bar", jschema.RuleASTNodeSourceManual)

	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeArray,
		Properties: &jschema.RuleASTNodes{},
		Items: []jschema.RuleASTNode{
			an,
			{
				JSONType:   jschema.JSONTypeString,
				Value:      "bar",
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			},
		},
		Source: jschema.RuleASTNodeSourceManual,
	}, l.ASTNode())
}
