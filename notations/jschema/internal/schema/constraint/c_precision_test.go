package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
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
