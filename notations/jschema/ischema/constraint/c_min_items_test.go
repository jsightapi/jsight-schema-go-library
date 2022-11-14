package constraint

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	schema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/json"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestNewMinItems(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cnstr := NewMinItems(bytes.NewBytes("10"))

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
				assert.PanicsWithError(t, `Invalid value of "minItems" constraint`, func() {
					NewMinItems(bytes.NewBytes(s))
				})
			})
		}
	})
}

func TestMinItems_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, MinItems{}, json.TypeArray)
}

func TestMinItems_Type(t *testing.T) {
	assert.Equal(t, MinItemsConstraintType, NewMinItems(bytes.NewBytes("1")).Type())
}

func TestMinItems_String(t *testing.T) {
	assert.Equal(t, "minItems: 1", NewMinItems(bytes.NewBytes("1")).String())
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
					NewMinItems(bytes.NewBytes("2")).ValidateTheArray(numberOfChildren)
				})
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `The number of array elements does not match the "minItems" rule`, func() {
			NewMinItems(bytes.NewBytes("2")).ValidateTheArray(1)
		})
	})
}

func TestMinItems_Value(t *testing.T) {
	assert.EqualValues(t, 2, NewMinItems(bytes.NewBytes("2")).Value())
}

func TestMinItems_ASTNode(t *testing.T) {
	assert.Equal(t, schema.RuleASTNode{
		TokenType:  schema.TokenTypeNumber,
		Value:      "1",
		Properties: &schema.RuleASTNodes{},
		Source:     schema.RuleASTNodeSourceManual,
	}, NewMinItems(bytes.NewBytes("1")).ASTNode())
}
