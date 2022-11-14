package constraint

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	schema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/json"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
)

func Test_NewAllOf(t *testing.T) {
	c := NewAllOf()
	assert.NotNil(t, c.schemaName)
}

func TestAllOf_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, AllOf{}, json.TypeObject)
}

func TestAllOf_Type(t *testing.T) {
	assert.Equal(t, AllOfConstraintType, NewAllOf().Type())
}

func TestAllOf_String(t *testing.T) {
	c := NewAllOf()
	c.Append(bytes.NewBytes(`"@foo"`))
	c.Append(bytes.NewBytes(`"@bar"`))
	c.Append(bytes.NewBytes(`"@fizz"`))
	c.Append(bytes.NewBytes(`"@buzz"`))

	assert.Equal(t, fmt.Sprintf("%s: @foo, @bar, @fizz, @buzz", AllOfConstraintType), c.String())
}

func TestAllOf_Append(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		c := NewAllOf()
		assert.Equal(t, []string{}, c.schemaName)

		c.Append(bytes.NewBytes(`"@foo"`))
		assert.Equal(t, []string{"@foo"}, c.schemaName)

		c.Append(bytes.NewBytes(`"@bar"`))
		assert.Equal(t, []string{"@foo", "@bar"}, c.schemaName)
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("not a string", func(t *testing.T) {
			assert.PanicsWithValue(t, errors.ErrUnacceptableValueInAllOfRule, func() {
				NewAllOf().Append(bytes.NewBytes("@foo"))
			})
		})

		t.Run("not a type", func(t *testing.T) {
			assert.PanicsWithError(t, `Invalid schema name (foo) in "allOf" rule`, func() {
				NewAllOf().Append(bytes.NewBytes(`"foo"`))
			})
		})
	})
}

func TestAllOf_SchemaNames(t *testing.T) {
	c := NewAllOf()
	c.Append(bytes.NewBytes(`"@foo"`))
	c.Append(bytes.NewBytes(`"@bar"`))
	c.Append(bytes.NewBytes(`"@fizz"`))
	c.Append(bytes.NewBytes(`"@buzz"`))

	assert.Equal(t, []string{"@foo", "@bar", "@fizz", "@buzz"}, c.SchemaNames())
}

func TestAllOf_ASTNode(t *testing.T) {
	cc := map[string]struct {
		setup    func() *AllOf
		expected schema.RuleASTNode
	}{
		"none": {
			setup: func() *AllOf {
				return NewAllOf()
			},

			expected: schema.RuleASTNode{
				TokenType:  schema.TokenTypeArray,
				Properties: &schema.RuleASTNodes{},
				Items:      []schema.RuleASTNode{},
				Source:     schema.RuleASTNodeSourceManual,
			},
		},

		"single": {
			setup: func() *AllOf {
				c := NewAllOf()
				c.Append(bytes.NewBytes(`"@foo"`))
				return c
			},

			expected: schema.RuleASTNode{
				TokenType:  schema.TokenTypeShortcut,
				Properties: &schema.RuleASTNodes{},
				Value:      "@foo",
				Source:     schema.RuleASTNodeSourceManual,
			},
		},

		"multiple": {
			setup: func() *AllOf {
				c := NewAllOf()
				c.Append(bytes.NewBytes(`"@foo"`))
				c.Append(bytes.NewBytes(`"@bar"`))
				c.Append(bytes.NewBytes(`"@fizz"`))
				c.Append(bytes.NewBytes(`"@buzz"`))
				return c
			},

			expected: schema.RuleASTNode{
				TokenType:  schema.TokenTypeArray,
				Properties: &schema.RuleASTNodes{},
				Items: []schema.RuleASTNode{
					{
						TokenType:  schema.TokenTypeShortcut,
						Value:      "@foo",
						Properties: &schema.RuleASTNodes{},
						Source:     schema.RuleASTNodeSourceManual,
					},
					{
						TokenType:  schema.TokenTypeShortcut,
						Value:      "@bar",
						Properties: &schema.RuleASTNodes{},
						Source:     schema.RuleASTNodeSourceManual,
					},
					{
						TokenType:  schema.TokenTypeShortcut,
						Value:      "@fizz",
						Properties: &schema.RuleASTNodes{},
						Source:     schema.RuleASTNodeSourceManual,
					},
					{
						TokenType:  schema.TokenTypeShortcut,
						Value:      "@buzz",
						Properties: &schema.RuleASTNodes{},
						Source:     schema.RuleASTNodeSourceManual,
					},
				},
				Source: schema.RuleASTNodeSourceManual,
			},
		},
	}

	for n, c := range cc {
		t.Run(n, func(t *testing.T) {
			assert.Equal(t, c.expected, c.setup().ASTNode())
		})
	}
}
