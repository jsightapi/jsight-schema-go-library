package constraint

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
)

func TestNewTypesList(t *testing.T) {
	c := NewTypesList(jschema.RuleASTNodeSourceGenerated)

	assert.NotNil(t, c.innerTypeNames)
	assert.Equal(t, jschema.RuleASTNodeSourceGenerated, c.source)
}

func TestTypesList_HasUserTypes(t *testing.T) {
	cc := []bool{
		false,
		true,
	}

	for _, given := range cc {
		t.Run(fmt.Sprintf("%t", given), func(t *testing.T) {
			assert.Equal(t, given, TypesList{hasUserTypes: given}.HasUserTypes())
		})
	}
}

func TestTypesList_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, TypesList{}, allJSONTypes...)
}

func TestTypesList_Type(t *testing.T) {
	assert.Equal(t, TypesListConstraintType, NewTypesList(jschema.RuleASTNodeSourceManual).Type())
}

func TestTypesList_String(t *testing.T) {
	c := NewTypesList(jschema.RuleASTNodeSourceManual)
	c.innerTypeNames = []string{"foo", "bar"}

	assert.Equal(t, "types: foo, bar", c.String())
}

func TestTypesList_AddName(t *testing.T) {
	c := NewTypesList(jschema.RuleASTNodeSourceManual)
	c.AddName("@foo", "bar", jschema.RuleASTNodeSourceGenerated)

	assert.Equal(t, []string{"@foo"}, c.innerTypeNames)
	assert.Equal(t, []string{"bar"}, c.typeNames)
	assert.Equal(t, []jschema.RuleASTNode{
		{
			TokenType:  jschema.TokenTypeString,
			Value:      "bar",
			Properties: &jschema.RuleASTNodes{},
			Source:     jschema.RuleASTNodeSourceGenerated,
		},
	}, c.elementASTNodes)
	assert.True(t, c.hasUserTypes)
}

func TestTypesList_AddNameWithASTNode(t *testing.T) {
	an := jschema.RuleASTNode{
		TokenType: jschema.TokenTypeString,
	}

	c := NewTypesList(jschema.RuleASTNodeSourceManual)
	c.AddNameWithASTNode("@foo", "bar", an)

	assert.Equal(t, []string{"@foo"}, c.innerTypeNames)
	assert.Equal(t, []string{"bar"}, c.typeNames)
	assert.Equal(t, []jschema.RuleASTNode{an}, c.elementASTNodes)
	assert.True(t, c.hasUserTypes)
}

func TestTypesList_Names(t *testing.T) {
	c := NewTypesList(jschema.RuleASTNodeSourceManual)
	c.innerTypeNames = []string{"foo", "bar"}

	assert.Equal(t, []string{"foo", "bar"}, c.Names())
}

func TestTypesList_Len(t *testing.T) {
	c := NewTypesList(jschema.RuleASTNodeSourceManual)
	c.innerTypeNames = []string{"foo", "bar"}

	assert.Equal(t, 2, c.Len())
}

func TestTypesList_ASTNode(t *testing.T) {
	l := NewTypesList(jschema.RuleASTNodeSourceManual)

	an := jschema.RuleASTNode{
		TokenType: jschema.TokenTypeObject,
		Properties: jschema.NewRuleASTNodes(
			map[string]jschema.RuleASTNode{
				"type": newRuleASTNode(jschema.TokenTypeString, "foo", jschema.RuleASTNodeSourceManual),
			},
			[]string{"type"},
		),
	}

	l.AddNameWithASTNode("foo", "foo", an)
	l.AddName("bar", "bar", jschema.RuleASTNodeSourceManual)

	assert.Equal(t, jschema.RuleASTNode{
		TokenType:  jschema.TokenTypeArray,
		Properties: &jschema.RuleASTNodes{},
		Items: []jschema.RuleASTNode{
			an,
			{
				TokenType:  jschema.TokenTypeString,
				Value:      "bar",
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			},
		},
		Source: jschema.RuleASTNodeSourceManual,
	}, l.ASTNode())
}

func TestTypesList_Source(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNodeSourceManual, NewTypesList(jschema.RuleASTNodeSourceManual).Source())
}
