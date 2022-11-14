package constraint

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	schema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/json"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestNewMaxItems(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cnstr := NewMaxItems(bytes.NewBytes("10"))

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
				assert.PanicsWithError(t, `Invalid value of "maxItems" constraint`, func() {
					NewMaxItems(bytes.NewBytes(s))
				})
			})
		}
	})
}

func TestMaxItems_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, MaxItems{}, json.TypeArray)
}

func TestMaxItems_Type(t *testing.T) {
	assert.Equal(t, MaxItemsConstraintType, NewMaxItems(bytes.NewBytes("1")).Type())
}

func TestMaxItems_String(t *testing.T) {
	assert.Equal(t, "maxItems: 1", NewMaxItems(bytes.NewBytes("1")).String())
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
					NewMaxItems(bytes.NewBytes("2")).ValidateTheArray(numberOfChildren)
				})
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `The number of array elements does not match the "maxItems" rule`, func() {
			NewMaxItems(bytes.NewBytes("2")).ValidateTheArray(3)
		})
	})
}

func TestMaxItems_Value(t *testing.T) {
	assert.EqualValues(t, 2, NewMaxItems(bytes.NewBytes("2")).Value())
}

func TestMaxItems_ASTNode(t *testing.T) {
	assert.Equal(t, schema.RuleASTNode{
		TokenType:  schema.TokenTypeNumber,
		Value:      "1",
		Properties: &schema.RuleASTNodes{},
		Source:     schema.RuleASTNodeSourceManual,
	}, NewMaxItems(bytes.NewBytes("1")).ASTNode())
}
