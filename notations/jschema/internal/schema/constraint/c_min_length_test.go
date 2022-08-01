package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestNewMinLength(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		ruleValue := bytes.Bytes("10")
		cnstr := NewMinLength(ruleValue)

		expectedNumber, err := json.NewNumber(ruleValue)
		require.NoError(t, err)

		assert.Equal(t, ruleValue, cnstr.rawValue)
		assert.Equal(t, expectedNumber, cnstr.value)
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Incorrect value "not an integer". Must be an integer.`, func() {
			NewMinLength([]byte("not an integer"))
		})
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
