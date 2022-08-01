package constraint

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
)

func TestNewOr(t *testing.T) {
	c := NewOr(jschema.RuleASTNodeSourceGenerated)
	assert.Equal(t, jschema.RuleASTNodeSourceGenerated, c.source)
}

func TestOr_IsGenerated(t *testing.T) {
	cc := map[jschema.RuleASTNodeSource]bool{
		jschema.RuleASTNodeSourceUnknown:   false,
		jschema.RuleASTNodeSourceManual:    false,
		jschema.RuleASTNodeSourceGenerated: true,
	}

	for source, expected := range cc {
		t.Run(strconv.Itoa(int(source)), func(t *testing.T) {
			assert.Equal(t, expected, NewOr(source).IsGenerated())
		})
	}
}

func TestOr_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, Or{}, allJSONTypes...)
}

func TestOr_Type(t *testing.T) {
	assert.Equal(t, OrConstraintType, NewOr(jschema.RuleASTNodeSourceGenerated).Type())
}

func TestOr_String(t *testing.T) {
	assert.Equal(t, "[ UNVERIFIABLE CONSTRAINT ] or", Or{}.String())
}

func TestOr_ASTNode(t *testing.T) {
	assert.Equal(t, newEmptyRuleASTNode(), NewOr(jschema.RuleASTNodeSourceManual).ASTNode())
}
