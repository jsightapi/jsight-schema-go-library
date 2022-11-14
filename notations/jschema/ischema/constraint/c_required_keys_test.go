package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	schema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/json"
)

func TestNewRequiredKeys(t *testing.T) {
	c := NewRequiredKeys()
	assert.NotNil(t, c.keys)
}

func TestRequiredKeys_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, RequiredKeys{}, json.TypeObject)
}

func TestRequiredKeys_Type(t *testing.T) {
	assert.Equal(t, RequiredKeysConstraintType, NewRequiredKeys().Type())
}

func TestRequiredKeys_String(t *testing.T) {
	c := NewRequiredKeys()
	c.AddKey("foo")
	c.AddKey("bar")

	assert.Equal(t, "required-keys: foo, bar", c.String())
}

func TestRequiredKeys_Keys(t *testing.T) {
	c := NewRequiredKeys()
	c.AddKey("foo")
	c.AddKey("bar")

	assert.Equal(t, []string{"foo", "bar"}, c.Keys())
}

func TestRequiredKeys_AddKey(t *testing.T) {
	c := NewRequiredKeys()
	assert.Equal(t, []string{}, c.keys)

	c.AddKey("foo")
	assert.Equal(t, []string{"foo"}, c.keys)

	c.AddKey("bar")
	assert.Equal(t, []string{"foo", "bar"}, c.keys)
}

func TestRequiredKeys_ASTNode(t *testing.T) {
	c := NewRequiredKeys()
	c.AddKey("foo")
	c.AddKey("bar")

	assert.Equal(t, schema.RuleASTNode{
		TokenType:  schema.TokenTypeArray,
		Properties: &schema.RuleASTNodes{},
		Items: []schema.RuleASTNode{
			{
				TokenType:  schema.TokenTypeString,
				Value:      "foo",
				Properties: &schema.RuleASTNodes{},
				Source:     schema.RuleASTNodeSourceManual,
			},
			{
				TokenType:  schema.TokenTypeString,
				Value:      "bar",
				Properties: &schema.RuleASTNodes{},
				Source:     schema.RuleASTNodeSourceManual,
			},
		},
		Source: schema.RuleASTNodeSourceManual,
	}, c.ASTNode())
}
