package constraint

import (
	"j/schema"
	"j/schema/bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinItems_Type(t *testing.T) {
	assert.Equal(t, MinItemsConstraintType, NewMinItems(bytes.Bytes("1")).Type())
}

func TestMinItems_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewMinItems(bytes.Bytes("1")).ASTNode())
}
