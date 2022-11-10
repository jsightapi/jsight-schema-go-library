package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestNewMaxLength(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cnstr := NewMaxLength(bytes.NewBytes("10"))

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
				assert.PanicsWithError(t, `Invalid value of "maxLength" constraint`, func() {
					NewMaxLength(bytes.NewBytes(s))
				})
			})
		}
	})
}

func TestMaxLength_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, MaxLength{}, json.TypeString)
}

func TestMaxLength_Type(t *testing.T) {
	assert.Equal(t, MaxLengthConstraintType, NewMaxLength(bytes.NewBytes("1")).Type())
}

func TestMaxLength_String(t *testing.T) {
	assert.Equal(t, "maxLength: 1", NewMaxLength(bytes.NewBytes("1")).String())
}

func TestMaxLength_Validate(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := []string{
			"",
			"foo",
			"0123456789",
		}

		for _, given := range cc {
			t.Run(given, func(t *testing.T) {
				assert.NotPanics(t, func() {
					NewMaxLength(bytes.NewBytes("10")).Validate(bytes.NewBytes(given))
				})
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Invalid string length for "maxLength" = "10" constraint`, func() {
			NewMaxLength(bytes.NewBytes("10")).Validate(bytes.NewBytes("0123456789A"))
		})
	})
}

func TestMaxLength_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		TokenType:  jschema.TokenTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewMaxLength(bytes.NewBytes("1")).ASTNode())
}

func TestMaxLength_Value(t *testing.T) {
	assert.Equal(t, uint(1), NewMaxLength(bytes.NewBytes("1")).Value())
}
