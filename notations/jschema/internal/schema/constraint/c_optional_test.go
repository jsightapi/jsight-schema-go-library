package constraint

import (
	"github.com/stretchr/testify/assert"
	jschema "j/schema"
	"j/schema/bytes"
	"strconv"
	"testing"
)

func TestOptional_Type(t *testing.T) {
	assert.Equal(t, OptionalConstraintType, NewOptional(bytes.Bytes("true")).Type())
}

func TestOptional_ASTNode(t *testing.T) {
	cc := []bool{true, false}

	for _, c := range cc {
		t.Run(strconv.FormatBool(c), func(t *testing.T) {
			assert.Equal(t, jschema.RuleASTNode{
				JSONType:   jschema.JSONTypeBoolean,
				Value:      strconv.FormatBool(c),
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			}, Optional{value: c}.ASTNode())
		})
	}
}
