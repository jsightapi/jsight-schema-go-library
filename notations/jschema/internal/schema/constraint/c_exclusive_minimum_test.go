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

func TestNewExclusiveMinimum(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]bool{
			"false": false,
			"true":  true,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				cnstr := NewExclusiveMinimum([]byte(given))
				assert.Equal(t, expected, cnstr.exclusive)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Invalid value of "exclusiveMinimum" constraint`, func() {
			NewExclusiveMinimum([]byte("42"))
		})
	})
}

func TestExclusiveMinimum_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, ExclusiveMinimum{}, json.TypeInteger, json.TypeFloat)
}

func TestExclusiveMinimum_Type(t *testing.T) {
	assert.Equal(t, ExclusiveMinimumConstraintType, NewExclusiveMinimum(bytes.Bytes("true")).Type())
}

func TestExclusiveMinimum_String(t *testing.T) {
	cc := map[bool]string{
		false: "[ UNVERIFIABLE CONSTRAINT ] exclusiveMinimum: false",
		true:  "[ UNVERIFIABLE CONSTRAINT ] exclusiveMinimum: true",
	}

	for given, expected := range cc {
		t.Run(expected, func(t *testing.T) {
			actual := ExclusiveMinimum{
				exclusive: given,
			}.
				String()

			assert.Equal(t, expected, actual)
		})
	}
}

func TestExclusiveMinimum_IsExclusive(t *testing.T) {
	cc := []bool{false, true}

	for _, expected := range cc {
		t.Run(fmt.Sprintf("%t", expected), func(t *testing.T) {
			actual := ExclusiveMinimum{
				exclusive: expected,
			}.
				IsExclusive()

			assert.Equal(t, expected, actual)
		})
	}
}

func TestExclusiveMinimum_ASTNode(t *testing.T) {
	cc := []bool{true, false}

	for _, c := range cc {
		t.Run(strconv.FormatBool(c), func(t *testing.T) {
			assert.Equal(t, jschema.RuleASTNode{
				TokenType:  jschema.TokenTypeBoolean,
				Value:      strconv.FormatBool(c),
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			}, ExclusiveMinimum{exclusive: c}.ASTNode())
		})
	}
}
