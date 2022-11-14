package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	schema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/json"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
)

func TestNewRegex(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		c := NewRegex(bytes.NewBytes(`"."`))

		assert.Equal(t, ".", c.expression)
		assert.NotNil(t, c.re)
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("invalid JSON", func(t *testing.T) {
			assert.PanicsWithError(t, "invalid character 'i' looking for beginning of value", func() {
				NewRegex(bytes.NewBytes("invalid"))
			})
		})

		t.Run("invalid expression", func(t *testing.T) {
			assert.PanicsWithValue(t, "regexp: Compile(`\\l`): error parsing regexp: invalid escape sequence: `\\l`", func() {
				NewRegex(bytes.NewBytes(`"\\l"`))
			})
		})
	})
}

func TestRegex_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, Regex{}, json.TypeString)
}

func TestRegexp_Type(t *testing.T) {
	assert.Equal(t, RegexConstraintType, NewRegex(bytes.NewBytes(`"."`)).Type())
}

func TestRegex_String(t *testing.T) {
	assert.Equal(t, `regex: .`, NewRegex(bytes.NewBytes(`"."`)).String())
}

func TestRegex_Validate(t *testing.T) {
	cnstr := NewRegex(bytes.NewBytes(`"foo-\\d"`))

	t.Run("valid", func(t *testing.T) {
		assert.NotPanics(t, func() {
			cnstr.Validate(bytes.NewBytes("foo-9"))
		})
	})

	t.Run("not valid", func(t *testing.T) {
		assert.PanicsWithValue(t, errors.ErrDoesNotMatchRegularExpression, func() {
			cnstr.Validate(bytes.NewBytes("foo-"))
		})
	})
}

func TestRegexp_ASTNode(t *testing.T) {
	assert.Equal(t, schema.RuleASTNode{
		TokenType:  schema.TokenTypeString,
		Value:      "foo",
		Properties: &schema.RuleASTNodes{},
		Source:     schema.RuleASTNodeSourceManual,
	}, Regex{expression: "foo"}.ASTNode())
}
