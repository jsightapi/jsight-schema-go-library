package constraint

import (
	jschema "j/schema"
	"j/schema/bytes"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullable_Type(t *testing.T) {
	assert.Equal(t, NullableConstraintType, NewNullable(bytes.Bytes("true")).Type())
}

func TestNullable_ASTNode(t *testing.T) {
	cc := []bool{true, false}

	for _, c := range cc {
		t.Run(strconv.FormatBool(c), func(t *testing.T) {
			assert.Equal(t, jschema.RuleASTNode{
				JSONType:   jschema.JSONTypeBoolean,
				Value:      strconv.FormatBool(c),
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			}, Nullable{value: c}.ASTNode())
		})
	}
}
