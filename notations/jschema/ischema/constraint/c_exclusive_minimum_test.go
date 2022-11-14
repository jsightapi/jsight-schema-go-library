package constraint

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	schema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/json"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestNewExclusiveMinimum(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]bool{
			"false": false,
			"true":  true,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				cnstr := NewExclusiveMinimum(bytes.NewBytes(given))
				assert.Equal(t, expected, cnstr.exclusive)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Invalid value of "exclusiveMinimum" constraint`, func() {
			NewExclusiveMinimum(bytes.NewBytes("42"))
		})
	})
}

func TestExclusiveMinimum_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, ExclusiveMinimum{}, json.TypeInteger, json.TypeFloat)
}

func TestExclusiveMinimum_Type(t *testing.T) {
	assert.Equal(t, ExclusiveMinimumConstraintType, NewExclusiveMinimum(bytes.NewBytes("true")).Type())
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
			assert.Equal(t, schema.RuleASTNode{
				TokenType:  schema.TokenTypeBoolean,
				Value:      strconv.FormatBool(c),
				Properties: &schema.RuleASTNodes{},
				Source:     schema.RuleASTNodeSourceManual,
			}, ExclusiveMinimum{exclusive: c}.ASTNode())
		})
	}
}
