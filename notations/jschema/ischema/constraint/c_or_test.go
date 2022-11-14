package constraint

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	schema "github.com/jsightapi/jsight-schema-go-library"
)

func TestNewOr(t *testing.T) {
	c := NewOr(schema.RuleASTNodeSourceGenerated)
	assert.Equal(t, schema.RuleASTNodeSourceGenerated, c.source)
}

func TestOr_IsGenerated(t *testing.T) {
	cc := map[schema.RuleASTNodeSource]bool{
		schema.RuleASTNodeSourceUnknown:   false,
		schema.RuleASTNodeSourceManual:    false,
		schema.RuleASTNodeSourceGenerated: true,
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
	assert.Equal(t, OrConstraintType, NewOr(schema.RuleASTNodeSourceGenerated).Type())
}

func TestOr_String(t *testing.T) {
	assert.Equal(t, "[ UNVERIFIABLE CONSTRAINT ] or", Or{}.String())
}

func TestOr_ASTNode(t *testing.T) {
	assert.Equal(t, newEmptyRuleASTNode(), NewOr(schema.RuleASTNodeSourceManual).ASTNode())
}
