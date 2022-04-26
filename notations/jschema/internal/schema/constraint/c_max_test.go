package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestMax_Type(t *testing.T) {
	assert.Equal(t, MaxConstraintType, NewMax(bytes.Bytes("1")).Type())
}

func TestMax_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewMax(bytes.Bytes("1")).ASTNode())
}
