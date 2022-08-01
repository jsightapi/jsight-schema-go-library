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

func TestNewMaxItems(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		ruleValue := bytes.Bytes("10")
		cnstr := NewMaxItems(ruleValue)

		expectedNumber, err := json.NewNumber(ruleValue)
		require.NoError(t, err)

		assert.Equal(t, ruleValue, cnstr.rawValue)
		assert.Equal(t, expectedNumber, cnstr.value)
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Incorrect value "not an integer". Must be an integer.`, func() {
			NewMaxItems([]byte("not an integer"))
		})
	})
}

func TestMaxItems_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, MaxItems{}, json.TypeArray)
}

func TestMaxItems_Type(t *testing.T) {
	assert.Equal(t, MaxItemsConstraintType, NewMaxItems(bytes.Bytes("1")).Type())
}

func TestMaxItems_String(t *testing.T) {
	assert.Equal(t, "maxItems: 1", NewMaxItems([]byte("1")).String())
}

func TestMaxItems_ValidateTheArray(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := []uint{
			1,
			2,
		}

		for _, numberOfChildren := range cc {
			t.Run(fmt.Sprintf("%d", numberOfChildren), func(t *testing.T) {
				assert.NotPanics(t, func() {
					NewMaxItems([]byte("2")).ValidateTheArray(numberOfChildren)
				})
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `The number of array elements does not match the "maxItems" rule`, func() {
			NewMaxItems([]byte("2")).ValidateTheArray(3)
		})
	})
}

func TestMaxItems_Value(t *testing.T) {
	given := []byte("2")
	expected, err := json.NewNumber(given)
	require.NoError(t, err)

	assert.Equal(t, expected, NewMaxItems(given).Value())
}

func TestMaxItems_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewMaxItems(bytes.Bytes("1")).ASTNode())
}
