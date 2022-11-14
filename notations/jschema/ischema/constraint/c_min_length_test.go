package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	schema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/json"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestNewMinLength(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cnstr := NewMinLength(bytes.NewBytes("10"))

		assert.EqualValues(t, 10, cnstr.value)
	})

	t.Run("negative", func(t *testing.T) {
		ss := []string{
			"not a number",
			"3.14",
			"-12",
		}

		for _, s := range ss {
			t.Run(s, func(t *testing.T) {
				assert.PanicsWithError(t, `Invalid value of "minLength" constraint`, func() {
					NewMinLength(bytes.NewBytes(s))
				})
			})
		}
	})
}

func TestMinLength_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, MinLength{}, json.TypeString)
}

func TestMinLength_Type(t *testing.T) {
	assert.Equal(t, MinLengthConstraintType, NewMinLength(bytes.NewBytes("1")).Type())
}

func TestMinLength_String(t *testing.T) {
	assert.Equal(t, "minLength: 1", NewMinLength(bytes.NewBytes("1")).String())
}

func TestMinLength_Validate(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := []string{
			"0123456789",
			"0123456789A",
			"0123456789AB",
		}

		for _, given := range cc {
			t.Run(given, func(t *testing.T) {
				assert.NotPanics(t, func() {
					NewMinLength(bytes.NewBytes("10")).Validate(bytes.NewBytes(given))
				})
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Invalid string length for "minLength" = "10" constraint`, func() {
			NewMinLength(bytes.NewBytes("10")).Validate(bytes.NewBytes("012345678"))
		})
	})
}

func TestMinLength_ASTNode(t *testing.T) {
	assert.Equal(t, schema.RuleASTNode{
		TokenType:  schema.TokenTypeNumber,
		Value:      "1",
		Properties: &schema.RuleASTNodes{},
		Source:     schema.RuleASTNodeSourceManual,
	}, NewMinLength(bytes.NewBytes("1")).ASTNode())
}

func TestMinLength_Value(t *testing.T) {
	assert.Equal(t, uint(1), NewMinLength(bytes.NewBytes("1")).Value())
}
