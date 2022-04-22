package constraint

import (
	"j/schema"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOr_Type(t *testing.T) {
	assert.Equal(t, OrConstraintType, NewOr(jschema.RuleASTNodeSourceGenerated).Type())
}

func TestOr_ASTNode(t *testing.T) {
	assert.Equal(t, newEmptyRuleASTNode(), NewOr(jschema.RuleASTNodeSourceManual).ASTNode())
}
