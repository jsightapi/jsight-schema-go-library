package constraint

import (
	jschema "j/schema"
	"j/schema/bytes"
	"testing"

	"github.com/stretchr/testify/assert"
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
