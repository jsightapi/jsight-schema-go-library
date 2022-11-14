package constraint

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	schema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestNewNullable(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]bool{
			"true":  true,
			"false": false,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				c := NewNullable(bytes.NewBytes(given))
				assert.Equal(t, expected, c.value)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Invalid value of "nullable" constraint`, func() {
			NewNullable(bytes.NewBytes("foo"))
		})
	})
}

func TestNullable_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, Nullable{}, allJSONTypes...)
}

func TestNullable_Type(t *testing.T) {
	assert.Equal(t, NullableConstraintType, NewNullable(bytes.NewBytes("true")).Type())
}

func TestNullable_String(t *testing.T) {
	cc := map[string]string{
		"false": "nullable: false",
		"true":  "nullable: true",
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			assert.Equal(t, expected, NewNullable(bytes.NewBytes(given)).String())
		})
	}
}

func TestNullable_Bool(t *testing.T) {
	cc := map[string]bool{
		"false": false,
		"true":  true,
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			assert.Equal(t, expected, NewNullable(bytes.NewBytes(given)).Bool())
		})
	}
}

func TestNullable_ASTNode(t *testing.T) {
	cc := []bool{true, false}

	for _, c := range cc {
		t.Run(strconv.FormatBool(c), func(t *testing.T) {
			assert.Equal(t, schema.RuleASTNode{
				TokenType:  schema.TokenTypeBoolean,
				Value:      strconv.FormatBool(c),
				Properties: &schema.RuleASTNodes{},
				Source:     schema.RuleASTNodeSourceManual,
			}, Nullable{value: c}.ASTNode())
		})
	}
}
