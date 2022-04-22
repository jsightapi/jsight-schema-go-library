package constraint

import (
	jschema "j/schema"
	"j/schema/bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrecision_Type(t *testing.T) {
	assert.Equal(t, PrecisionConstraintType, NewPrecision(bytes.Bytes("1")).Type())
}

func TestPrecision_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewPrecision(bytes.Bytes("1")).ASTNode())
}
