package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAny(t *testing.T) {
	assert.NotNil(t, NewAny())
}

func TestAnyConstraint_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, NewAny(), allJSONTypes...)
}

func TestAnyConstraint_Type(t *testing.T) {
	assert.Equal(t, AnyConstraintType, NewAny().Type())
}

func TestAnyConstraint_String(t *testing.T) {
	assert.Equal(t, AnyConstraintType.String(), NewAny().String())
}

func TestAnyConstraint_ASTNode(t *testing.T) {
	assert.Equal(t, newEmptyRuleASTNode(), NewAny().ASTNode())
}
