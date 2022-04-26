package constraint

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
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
