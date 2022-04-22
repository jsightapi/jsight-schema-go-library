package constraint

import (
	"j/schema"
	"j/schema/bytes"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExclusiveMaximum_Type(t *testing.T) {
	assert.Equal(t, ExclusiveMaximumConstraintType, NewExclusiveMaximum(bytes.Bytes("true")).Type())
}

func TestExclusiveMaximum_ASTNode(t *testing.T) {
	cc := []bool{true, false}

	for _, c := range cc {
		t.Run(strconv.FormatBool(c), func(t *testing.T) {
			assert.Equal(t, jschema.RuleASTNode{
				JSONType:   jschema.JSONTypeBoolean,
				Value:      strconv.FormatBool(c),
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			}, ExclusiveMaximum{exclusive: c}.ASTNode())
		})
	}
}
