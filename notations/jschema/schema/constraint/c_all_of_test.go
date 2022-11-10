package constraint

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
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
		expected jschema.RuleASTNode
	}{
		"none": {
			setup: func() *AllOf {
				return NewAllOf()
			},

			expected: jschema.RuleASTNode{
				TokenType:  jschema.TokenTypeArray,
				Properties: &jschema.RuleASTNodes{},
				Items:      []jschema.RuleASTNode{},
				Source:     jschema.RuleASTNodeSourceManual,
			},
		},

		"single": {
			setup: func() *AllOf {
				c := NewAllOf()
				c.Append(bytes.NewBytes(`"@foo"`))
				return c
			},

			expected: jschema.RuleASTNode{
				TokenType:  jschema.TokenTypeShortcut,
				Properties: &jschema.RuleASTNodes{},
				Value:      "@foo",
				Source:     jschema.RuleASTNodeSourceManual,
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

			expected: jschema.RuleASTNode{
				TokenType:  jschema.TokenTypeArray,
				Properties: &jschema.RuleASTNodes{},
				Items: []jschema.RuleASTNode{
					{
						TokenType:  jschema.TokenTypeShortcut,
						Value:      "@foo",
						Properties: &jschema.RuleASTNodes{},
						Source:     jschema.RuleASTNodeSourceManual,
					},
					{
						TokenType:  jschema.TokenTypeShortcut,
						Value:      "@bar",
						Properties: &jschema.RuleASTNodes{},
						Source:     jschema.RuleASTNodeSourceManual,
					},
					{
						TokenType:  jschema.TokenTypeShortcut,
						Value:      "@fizz",
						Properties: &jschema.RuleASTNodes{},
						Source:     jschema.RuleASTNodeSourceManual,
					},
					{
						TokenType:  jschema.TokenTypeShortcut,
						Value:      "@buzz",
						Properties: &jschema.RuleASTNodes{},
						Source:     jschema.RuleASTNodeSourceManual,
					},
				},
				Source: jschema.RuleASTNodeSourceManual,
			},
		},
	}

	for n, c := range cc {
		t.Run(n, func(t *testing.T) {
			assert.Equal(t, c.expected, c.setup().ASTNode())
		})
	}
}
