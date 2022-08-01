package constraint

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestNewExclusiveMaximum(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]bool{
			"false": false,
			"true":  true,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				cnstr := NewExclusiveMaximum([]byte(given))
				assert.Equal(t, expected, cnstr.exclusive)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Invalid value of "exclusiveMaximum" constraint`, func() {
			NewExclusiveMaximum([]byte("42"))
		})
	})
}

func TestExclusiveMaximum_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, ExclusiveMaximum{}, json.TypeInteger, json.TypeFloat)
}

func TestExclusiveMaximum_Type(t *testing.T) {
	assert.Equal(t, ExclusiveMaximumConstraintType, NewExclusiveMaximum(bytes.Bytes("true")).Type())
}

func TestExclusiveMaximum_String(t *testing.T) {
	cc := map[bool]string{
		false: "[ UNVERIFIABLE CONSTRAINT ] exclusiveMaximum: false",
		true:  "[ UNVERIFIABLE CONSTRAINT ] exclusiveMaximum: true",
	}

	for given, expected := range cc {
		t.Run(expected, func(t *testing.T) {
			actual := ExclusiveMaximum{
				exclusive: given,
			}.
				String()

			assert.Equal(t, expected, actual)
		})
	}
}

func TestExclusiveMaximum_IsExclusive(t *testing.T) {
	cc := []bool{false, true}

	for _, expected := range cc {
		t.Run(fmt.Sprintf("%t", expected), func(t *testing.T) {
			actual := ExclusiveMaximum{
				exclusive: expected,
			}.
				IsExclusive()

			assert.Equal(t, expected, actual)
		})
	}
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
