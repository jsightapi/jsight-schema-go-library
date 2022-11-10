package constraint

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestNewType(t *testing.T) {
	ruleValue := bytes.NewBytes("@foo")
	c := NewType(ruleValue, jschema.RuleASTNodeSourceGenerated)

	assert.Equal(t, ruleValue, c.value)
	assert.Equal(t, jschema.RuleASTNodeSourceGenerated, c.source)
}

func TestTypeConstraint_IsGenerated(t *testing.T) {
	cc := map[jschema.RuleASTNodeSource]bool{
		jschema.RuleASTNodeSourceUnknown:   false,
		jschema.RuleASTNodeSourceManual:    false,
		jschema.RuleASTNodeSourceGenerated: true,
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
		NewType(bytes.NewBytes("foo"), jschema.RuleASTNodeSourceGenerated).Type(),
	)
}

func TestTypeConstraint_String(t *testing.T) {
	assert.Equal(t, "type: @foo", NewType(bytes.NewBytes("@foo"), jschema.RuleASTNodeSourceGenerated).String())
}

func TestTypeConstraint_Bytes(t *testing.T) {
	ruleValue := bytes.NewBytes("@foo")
	c := NewType(ruleValue, jschema.RuleASTNodeSourceManual)

	assert.Equal(t, ruleValue, c.Bytes())
}

func TestTypeConstraint_ASTNode(t *testing.T) {
	cc := map[string]jschema.RuleASTNode{
		"foo": {
			TokenType:  jschema.TokenTypeString,
			Value:      "foo",
			Properties: &jschema.RuleASTNodes{},
			Source:     jschema.RuleASTNodeSourceGenerated,
		},

		"@foo": {
			TokenType:  jschema.TokenTypeShortcut,
			Value:      "@foo",
			Properties: &jschema.RuleASTNodes{},
			Source:     jschema.RuleASTNodeSourceGenerated,
		},
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			assert.Equal(
				t,
				expected,
				NewType(bytes.NewBytes(given), jschema.RuleASTNodeSourceGenerated).ASTNode(),
			)
		})
	}
}

func TestTypeConstraint_Source(t *testing.T) {
	assert.Equal(
		t,
		jschema.RuleASTNodeSourceManual,
		NewType(bytes.NewBytes("@foo"), jschema.RuleASTNodeSourceManual).Source(),
	)
}
