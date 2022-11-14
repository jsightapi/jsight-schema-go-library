package constraint

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	schema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestNewType(t *testing.T) {
	ruleValue := bytes.NewBytes("@foo")
	c := NewType(ruleValue, schema.RuleASTNodeSourceGenerated)

	assert.Equal(t, ruleValue, c.value)
	assert.Equal(t, schema.RuleASTNodeSourceGenerated, c.source)
}

func TestTypeConstraint_IsGenerated(t *testing.T) {
	cc := map[schema.RuleASTNodeSource]bool{
		schema.RuleASTNodeSourceUnknown:   false,
		schema.RuleASTNodeSourceManual:    false,
		schema.RuleASTNodeSourceGenerated: true,
	}

	for source, expected := range cc {
		t.Run(strconv.Itoa(int(source)), func(t *testing.T) {
			assert.Equal(t, expected, NewType(bytes.NewBytes("@foo"), source).IsGenerated())
		})
	}
}

func TestTypeConstraint_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, TypeConstraint{}, allJSONTypes...)
}

func TestTypeConstraint_Type(t *testing.T) {
	assert.Equal(t,
		TypeConstraintType,
		NewType(bytes.NewBytes("foo"), schema.RuleASTNodeSourceGenerated).Type(),
	)
}

func TestTypeConstraint_String(t *testing.T) {
	assert.Equal(t, "type: @foo", NewType(bytes.NewBytes("@foo"), schema.RuleASTNodeSourceGenerated).String())
}

func TestTypeConstraint_Bytes(t *testing.T) {
	ruleValue := bytes.NewBytes("@foo")
	c := NewType(ruleValue, schema.RuleASTNodeSourceManual)

	assert.Equal(t, ruleValue, c.Bytes())
}

func TestTypeConstraint_ASTNode(t *testing.T) {
	cc := map[string]schema.RuleASTNode{
		"foo": {
			TokenType:  schema.TokenTypeString,
			Value:      "foo",
			Properties: &schema.RuleASTNodes{},
			Source:     schema.RuleASTNodeSourceGenerated,
		},

		"@foo": {
			TokenType:  schema.TokenTypeShortcut,
			Value:      "@foo",
			Properties: &schema.RuleASTNodes{},
			Source:     schema.RuleASTNodeSourceGenerated,
		},
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			assert.Equal(
				t,
				expected,
				NewType(bytes.NewBytes(given), schema.RuleASTNodeSourceGenerated).ASTNode(),
			)
		})
	}
}

func TestTypeConstraint_Source(t *testing.T) {
	assert.Equal(
		t,
		schema.RuleASTNodeSourceManual,
		NewType(bytes.NewBytes("@foo"), schema.RuleASTNodeSourceManual).Source(),
	)
}
