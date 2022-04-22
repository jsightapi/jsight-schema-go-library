package constraint

import (
	"j/schema"
	"j/schema/bytes"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExclusiveMinimum_Type(t *testing.T) {
	assert.Equal(t, ExclusiveMinimumConstraintType, NewExclusiveMinimum(bytes.Bytes("true")).Type())
}

func TestExclusiveMinimum_ASTNode(t *testing.T) {
	cc := []bool{true, false}

	for _, c := range cc {
		t.Run(strconv.FormatBool(c), func(t *testing.T) {
			assert.Equal(t, jschema.RuleASTNode{
				JSONType:   jschema.JSONTypeBoolean,
				Value:      strconv.FormatBool(c),
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			}, ExclusiveMinimum{exclusive: c}.ASTNode())
		})
	}
}
