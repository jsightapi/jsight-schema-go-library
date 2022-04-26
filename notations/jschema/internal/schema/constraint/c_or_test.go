package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
)

func TestOr_Type(t *testing.T) {
	assert.Equal(t, OrConstraintType, NewOr(jschema.RuleASTNodeSourceGenerated).Type())
}

func TestOr_ASTNode(t *testing.T) {
	assert.Equal(t, newEmptyRuleASTNode(), NewOr(jschema.RuleASTNodeSourceManual).ASTNode())
}
