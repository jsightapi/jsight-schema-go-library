package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestNewMinLength(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cnstr := NewMinLength([]byte("10"))

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
					NewMinLength([]byte(s))
				})
			})
		}
	})
}

func TestMinLength_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, MinLength{}, json.TypeString)
}

func TestMinLength_Type(t *testing.T) {
	assert.Equal(t, MinLengthConstraintType, NewMinLength(bytes.Bytes("1")).Type())
}

func TestMinLength_String(t *testing.T) {
	assert.Equal(t, "minLength: 1", NewMinLength([]byte("1")).String())
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
					NewMinLength([]byte("10")).Validate([]byte(given))
				})
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Invalid string length for "minLength" = "10" constraint`, func() {
			NewMinLength([]byte("10")).Validate([]byte("012345678"))
		})
	})
}

func TestMinLength_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewMinLength(bytes.Bytes("1")).ASTNode())
}
