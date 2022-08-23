package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestNewRegex(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		c := NewRegex([]byte(`"."`))

		assert.Equal(t, ".", c.expression)
		assert.NotNil(t, c.re)
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("invalid JSON", func(t *testing.T) {
			assert.PanicsWithError(t, "invalid character 'i' looking for beginning of value", func() {
				NewRegex([]byte("invalid"))
			})
		})

		t.Run("invalid expression", func(t *testing.T) {
			assert.PanicsWithValue(t, "regexp: Compile(`\\l`): error parsing regexp: invalid escape sequence: `\\l`", func() {
				NewRegex([]byte(`"\\l"`))
			})
		})
	})
}

func TestRegex_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, Regex{}, json.TypeString)
}

func TestRegexp_Type(t *testing.T) {
	assert.Equal(t, RegexConstraintType, NewRegex(bytes.Bytes(`"."`)).Type())
}

func TestRegex_String(t *testing.T) {
	assert.Equal(t, `regex: .`, NewRegex([]byte(`"."`)).String())
}

func TestRegex_Validate(t *testing.T) {
	cnstr := NewRegex([]byte(`"foo-\\d"`))

	t.Run("valid", func(t *testing.T) {
		assert.NotPanics(t, func() {
			cnstr.Validate([]byte("foo-9"))
		})
	})

	t.Run("not valid", func(t *testing.T) {
		assert.PanicsWithValue(t, errors.ErrDoesNotMatchRegularExpression, func() {
			cnstr.Validate([]byte("foo-"))
		})
	})
}

func TestRegexp_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		TokenType:  jschema.TokenTypeString,
		Value:      "foo",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, Regex{expression: "foo"}.ASTNode())
}
