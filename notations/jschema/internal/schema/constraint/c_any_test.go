package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnyConstraint_Type(t *testing.T) {
	assert.Equal(t, AnyConstraintType, NewAny().Type())
}

func TestAnyConstraint_ASTNode(t *testing.T) {
	assert.Equal(t, newEmptyRuleASTNode(), NewAny().ASTNode())
}
