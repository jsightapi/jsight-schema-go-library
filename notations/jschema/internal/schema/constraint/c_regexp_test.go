package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestRegexp_Type(t *testing.T) {
	assert.Equal(t, RegexConstraintType, NewRegex(bytes.Bytes(`"."`)).Type())
}

func TestRegexp_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeString,
		Value:      "foo",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, Regex{expression: "foo"}.ASTNode())
}
