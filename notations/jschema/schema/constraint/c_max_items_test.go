package constraint

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestNewMaxItems(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cnstr := NewMaxItems([]byte("10"))

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
					NewMaxItems([]byte(s))
				})
			})
		}
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
	assert.EqualValues(t, 2, NewMaxItems([]byte("2")).Value())
}

func TestMaxItems_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		TokenType:  jschema.TokenTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewMaxItems(bytes.Bytes("1")).ASTNode())
}
