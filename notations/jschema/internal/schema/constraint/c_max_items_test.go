package constraint

import (
	"j/schema"
	"j/schema/bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMaxItems_Type(t *testing.T) {
	assert.Equal(t, MaxItemsConstraintType, NewMaxItems(bytes.Bytes("1")).Type())
}

func TestMaxItems_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewMaxItems(bytes.Bytes("1")).ASTNode())
}
