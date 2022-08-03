package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestNewMaxLength(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		ruleValue := bytes.Bytes("10")
		cnstr := NewMaxLength(ruleValue)

		expectedNumber, err := json.NewNumber(ruleValue)
		require.NoError(t, err)

		assert.Equal(t, ruleValue, cnstr.rawValue)
		assert.Equal(t, expectedNumber, cnstr.value)
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Incorrect value "not an integer". Must be an integer.`, func() {
			NewMaxLength([]byte("not an integer"))
		})
	})
}

func TestMaxLength_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, MaxLength{}, json.TypeString)
}

func TestMaxLength_Type(t *testing.T) {
	assert.Equal(t, MaxLengthConstraintType, NewMaxLength(bytes.Bytes("1")).Type())
}

func TestMaxLength_String(t *testing.T) {
	assert.Equal(t, "maxLength: 1", NewMaxLength([]byte("1")).String())
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
					NewMaxLength([]byte("10")).Validate([]byte(given))
				})
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Invalid string length for "maxLength" = "10" constraint`, func() {
			NewMaxLength([]byte("10")).Validate([]byte("0123456789A"))
		})
	})
}

func TestMaxLength_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewMaxLength(bytes.Bytes("1")).ASTNode())
}
