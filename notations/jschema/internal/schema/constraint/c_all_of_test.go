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
	cc := map[json.Type]bool{
		json.TypeObject:  true,
		json.TypeArray:   false,
		json.TypeString:  false,
		json.TypeInteger: false,
		json.TypeFloat:   false,
		json.TypeBoolean: false,
		json.TypeNull:    false,
	}

	for jsonType, expected := range cc {
		t.Run(jsonType.String(), func(t *testing.T) {
			assert.Equal(t, expected, AllOf{}.IsJsonTypeCompatible(jsonType))
		})
	}
}

func TestAllOf_Type(t *testing.T) {
	assert.Equal(t, AllOfConstraintType, NewAllOf().Type())
}

func TestAllOf_String(t *testing.T) {
	c := NewAllOf()
	c.Append(bytes.Bytes(`"@foo"`))
	c.Append(bytes.Bytes(`"@bar"`))
	c.Append(bytes.Bytes(`"@fizz"`))
	c.Append(bytes.Bytes(`"@buzz"`))

	assert.Equal(t, fmt.Sprintf("%s: @foo, @bar, @fizz, @buzz", AllOfConstraintType), c.String())
}

func TestAllOf_Append(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		c := NewAllOf()
		assert.Equal(t, []string{}, c.schemaName)

		c.Append(bytes.Bytes(`"@foo"`))
		assert.Equal(t, []string{"@foo"}, c.schemaName)

		c.Append(bytes.Bytes(`"@bar"`))
		assert.Equal(t, []string{"@foo", "@bar"}, c.schemaName)
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("not a string", func(t *testing.T) {
			assert.PanicsWithValue(t, errors.ErrUnacceptableValueInAllOfRule, func() {
				NewAllOf().Append(bytes.Bytes("@foo"))
			})
		})

		t.Run("not a type", func(t *testing.T) {
			assert.Panics(t, func() {
				NewAllOf().Append(bytes.Bytes(`"foo"`))
			})
		})
	})
}

func TestAllOf_SchemaNames(t *testing.T) {
	c := NewAllOf()
	c.Append(bytes.Bytes(`"@foo"`))
	c.Append(bytes.Bytes(`"@bar"`))
	c.Append(bytes.Bytes(`"@fizz"`))
	c.Append(bytes.Bytes(`"@buzz"`))

	assert.Equal(t, []string{"@foo", "@bar", "@fizz", "@buzz"}, c.SchemaNames())
}

func TestAllOf_ASTNode(t *testing.T) {
	c := NewAllOf()
	c.Append(bytes.Bytes(`"@foo"`))
	c.Append(bytes.Bytes(`"@bar"`))
	c.Append(bytes.Bytes(`"@fizz"`))
	c.Append(bytes.Bytes(`"@buzz"`))

	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeArray,
		Properties: &jschema.RuleASTNodes{},
		Items: []jschema.RuleASTNode{
			{
				JSONType:   jschema.JSONTypeString,
				Value:      "@foo",
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			},
			{
				JSONType:   jschema.JSONTypeString,
				Value:      "@bar",
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			},
			{
				JSONType:   jschema.JSONTypeString,
				Value:      "@fizz",
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			},
			{
				JSONType:   jschema.JSONTypeString,
				Value:      "@buzz",
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			},
		},
		Source: jschema.RuleASTNodeSourceManual,
	}, c.ASTNode())
}
