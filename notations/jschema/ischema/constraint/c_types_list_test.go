package constraint

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	schema "github.com/jsightapi/jsight-schema-go-library"
)

func TestNewTypesList(t *testing.T) {
	c := NewTypesList(schema.RuleASTNodeSourceGenerated)

	assert.NotNil(t, c.innerTypeNames)
	assert.Equal(t, schema.RuleASTNodeSourceGenerated, c.source)
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
	assert.Equal(t, TypesListConstraintType, NewTypesList(schema.RuleASTNodeSourceManual).Type())
}

func TestTypesList_String(t *testing.T) {
	c := NewTypesList(schema.RuleASTNodeSourceManual)
	c.innerTypeNames = []string{"foo", "bar"}

	assert.Equal(t, "types: foo, bar", c.String())
}

func TestTypesList_AddName(t *testing.T) {
	c := NewTypesList(schema.RuleASTNodeSourceManual)
	c.AddName("@foo", "bar", schema.RuleASTNodeSourceGenerated)

	assert.Equal(t, []string{"@foo"}, c.innerTypeNames)
	assert.Equal(t, []string{"bar"}, c.typeNames)
	assert.Equal(t, []schema.RuleASTNode{
		{
			TokenType:  schema.TokenTypeString,
			Value:      "bar",
			Properties: &schema.RuleASTNodes{},
			Source:     schema.RuleASTNodeSourceGenerated,
		},
	}, c.elementASTNodes)
	assert.True(t, c.hasUserTypes)
}

func TestTypesList_AddNameWithASTNode(t *testing.T) {
	an := schema.RuleASTNode{
		TokenType: schema.TokenTypeString,
	}

	c := NewTypesList(schema.RuleASTNodeSourceManual)
	c.AddNameWithASTNode("@foo", "bar", an)

	assert.Equal(t, []string{"@foo"}, c.innerTypeNames)
	assert.Equal(t, []string{"bar"}, c.typeNames)
	assert.Equal(t, []schema.RuleASTNode{an}, c.elementASTNodes)
	assert.True(t, c.hasUserTypes)
}

func TestTypesList_Names(t *testing.T) {
	c := NewTypesList(schema.RuleASTNodeSourceManual)
	c.innerTypeNames = []string{"foo", "bar"}

	assert.Equal(t, []string{"foo", "bar"}, c.Names())
}

func TestTypesList_Len(t *testing.T) {
	c := NewTypesList(schema.RuleASTNodeSourceManual)
	c.innerTypeNames = []string{"foo", "bar"}

	assert.Equal(t, 2, c.Len())
}

func TestTypesList_ASTNode(t *testing.T) {
	l := NewTypesList(schema.RuleASTNodeSourceManual)

	an := schema.RuleASTNode{
		TokenType: schema.TokenTypeObject,
		Properties: schema.NewRuleASTNodes(
			map[string]schema.RuleASTNode{
				"type": newRuleASTNode(schema.TokenTypeString, "foo", schema.RuleASTNodeSourceManual),
			},
			[]string{"type"},
		),
	}

	l.AddNameWithASTNode("foo", "foo", an)
	l.AddName("bar", "bar", schema.RuleASTNodeSourceManual)

	assert.Equal(t, schema.RuleASTNode{
		TokenType:  schema.TokenTypeArray,
		Properties: &schema.RuleASTNodes{},
		Items: []schema.RuleASTNode{
			an,
			{
				TokenType:  schema.TokenTypeString,
				Value:      "bar",
				Properties: &schema.RuleASTNodes{},
				Source:     schema.RuleASTNodeSourceManual,
			},
		},
		Source: schema.RuleASTNodeSourceManual,
	}, l.ASTNode())
}

func TestTypesList_Source(t *testing.T) {
	assert.Equal(t, schema.RuleASTNodeSourceManual, NewTypesList(schema.RuleASTNodeSourceManual).Source())
}
