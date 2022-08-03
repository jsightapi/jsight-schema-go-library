package constraint

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestNewMinItems(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		ruleValue := bytes.Bytes("10")
		cnstr := NewMinItems(ruleValue)

		expectedNumber, err := json.NewNumber(ruleValue)
		require.NoError(t, err)

		assert.Equal(t, ruleValue, cnstr.rawValue)
		assert.Equal(t, expectedNumber, cnstr.value)
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Incorrect value "not an integer". Must be an integer.`, func() {
			NewMinItems([]byte("not an integer"))
		})
	})
}

func TestMinItems_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, MinItems{}, json.TypeArray)
}

func TestMinItems_Type(t *testing.T) {
	assert.Equal(t, MinItemsConstraintType, NewMinItems(bytes.Bytes("1")).Type())
}

func TestMinItems_String(t *testing.T) {
	assert.Equal(t, "minItems: 1", NewMinItems([]byte("1")).String())
}

func TestMinItems_ValidateTheArray(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := []uint{
			2,
			3,
		}

		for _, numberOfChildren := range cc {
			t.Run(fmt.Sprintf("%d", numberOfChildren), func(t *testing.T) {
				assert.NotPanics(t, func() {
					NewMinItems([]byte("2")).ValidateTheArray(numberOfChildren)
				})
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `The number of array elements does not match the "minItems" rule`, func() {
			NewMinItems([]byte("2")).ValidateTheArray(1)
		})
	})
}

func TestMinItems_Value(t *testing.T) {
	given := []byte("2")
	expected, err := json.NewNumber(given)
	require.NoError(t, err)

	assert.Equal(t, expected, NewMinItems(given).Value())
}

func TestMinItems_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewMinItems(bytes.Bytes("1")).ASTNode())
}
