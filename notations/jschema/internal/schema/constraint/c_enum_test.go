package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestNewEnum(t *testing.T) {
	c := NewEnum()
	assert.NotNil(t, c.items)
}

func TestEnum_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(
		t,
		Enum{},
		json.TypeString,
		json.TypeBoolean,
		json.TypeInteger,
		json.TypeFloat,
		json.TypeNull,
		json.TypeMixed,
	)
}

func TestEnum_Type(t *testing.T) {
	assert.Equal(t, EnumConstraintType, NewEnum().Type())
}

func TestEnum_String(t *testing.T) {
	actual := Enum{items: []enumItem{
		{"", enumItemValue{value: `foo`, jsonType: json.TypeString}},
		{"", enumItemValue{value: `bar`, jsonType: json.TypeString}},
		{"", enumItemValue{value: `fizz`, jsonType: json.TypeString}},
		{"", enumItemValue{value: `buzz`, jsonType: json.TypeString}},
		{"", enumItemValue{value: `123`, jsonType: json.TypeInteger}},
		{"", enumItemValue{value: `45.67`, jsonType: json.TypeFloat}},
		{"", enumItemValue{value: `true`, jsonType: json.TypeBoolean}},
		{"", enumItemValue{value: `false`, jsonType: json.TypeBoolean}},
		{"", enumItemValue{value: `null`, jsonType: json.TypeNull}},
	}}.
		String()

	assert.Equal(t, `enum: ["foo", "bar", "fizz", "buzz", 123, 45.67, true, false, null]`, actual)
}

func TestEnum_Append(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		c := NewEnum()
		assert.Equal(t, []enumItem{}, c.items)

		c.Append(bytes.Bytes(`"foo"`))
		assert.Equal(t, []enumItem{
			{"", enumItemValue{value: `foo`, jsonType: json.TypeString}},
		}, c.items)

		c.Append(bytes.Bytes(`"bar"`))
		assert.Equal(t, []enumItem{
			{"", enumItemValue{value: `foo`, jsonType: json.TypeString}},
			{"", enumItemValue{value: `bar`, jsonType: json.TypeString}},
		}, c.items)

		c.Append(bytes.Bytes(`"FoO"`))
		assert.Equal(t, []enumItem{
			{"", enumItemValue{value: `foo`, jsonType: json.TypeString}},
			{"", enumItemValue{value: `bar`, jsonType: json.TypeString}},
			{"", enumItemValue{value: `FoO`, jsonType: json.TypeString}},
		}, c.items)
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]string{
			`"foo" value duplicates in "enum"`:  `"foo"`,
			` "foo" value duplicates in "enum"`: ` "foo"`,
			`"foo"  value duplicates in "enum"`: `"foo" `,
		}

		for expected, given := range cc {
			t.Run(expected, func(t *testing.T) {
				assert.PanicsWithError(t, expected, func() {
					c := NewEnum()
					c.Append(bytes.Bytes(`"foo"`))
					c.Append(bytes.Bytes(given))
				})
			})
		}
	})
}

func TestEnum_SetComment(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		e := &Enum{
			items: []enumItem{
				{"", enumItemValue{value: `foo`, jsonType: json.TypeString}},
				{"old bar comment", enumItemValue{value: `bar`, jsonType: json.TypeString}},
			},
		}

		e.SetComment(0, "foo comment")
		e.SetComment(1, "new bar comment")

		assert.Equal(t, []enumItem{
			{"foo comment", enumItemValue{value: `foo`, jsonType: json.TypeString}},
			{"new bar comment", enumItemValue{value: `bar`, jsonType: json.TypeString}},
		}, e.items)
	})

	t.Run("negative", func(t *testing.T) {
		assert.Panics(t, func() {
			(&Enum{}).SetComment(10, "panic")
		})
	})
}

func TestEnum_SetRuleName(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		e := &Enum{}

		e.SetRuleName("foo")

		assert.Equal(t, "foo", e.ruleName)
		assert.Equal(t, "foo", e.RuleName())
	})

	t.Run("negative", func(t *testing.T) {
		assert.Panics(t, func() {
			var e *Enum
			e.SetRuleName("panic")
		})
	})
}

func TestEnum_Validate(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		Enum{
			items: []enumItem{
				{"", enumItemValue{value: `foo`, jsonType: json.TypeString}},
				{"", enumItemValue{value: `bar`, jsonType: json.TypeString}},
			},
		}.
			Validate(bytes.Bytes(`"bar"`))
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithValue(t, errors.ErrDoesNotMatchAnyOfTheEnumValues, func() {
			Enum{
				items: []enumItem{
					{"", enumItemValue{value: `foo`, jsonType: json.TypeString}},
					{"", enumItemValue{value: `bar`, jsonType: json.TypeString}},
				},
			}.
				Validate(bytes.Bytes(`"fizz"`))
		})
	})
}

func TestEnum_ASTNode(t *testing.T) {
	t.Run("with rule name", func(t *testing.T) {
		e := Enum{
			ruleName: "@foo",
		}

		assert.Equal(t, jschema.RuleASTNode{
			TokenType:  jschema.TokenTypeShortcut,
			Value:      "@foo",
			Properties: &jschema.RuleASTNodes{},
			Source:     jschema.RuleASTNodeSourceManual,
		}, e.ASTNode())
	})

	t.Run("without rule name", func(t *testing.T) {
		e := Enum{
			items: []enumItem{
				{"", enumItemValue{value: `foo`, jsonType: json.TypeString}},
				{"foo", enumItemValue{value: `42`, jsonType: json.TypeInteger}},
				{"", enumItemValue{value: `3.14`, jsonType: json.TypeFloat}},
				{"", enumItemValue{value: `true`, jsonType: json.TypeBoolean}},
				{"bar", enumItemValue{value: `null`, jsonType: json.TypeNull}},
				{"", enumItemValue{value: `@foo`, jsonType: json.TypeMixed}},
			},
		}
		assert.Equal(t, jschema.RuleASTNode{
			TokenType:  jschema.TokenTypeArray,
			Properties: &jschema.RuleASTNodes{},
			Items: []jschema.RuleASTNode{
				{
					TokenType:  jschema.TokenTypeString,
					Value:      "foo",
					Properties: &jschema.RuleASTNodes{},
					Source:     jschema.RuleASTNodeSourceManual,
				},
				{
					TokenType:  jschema.TokenTypeNumber,
					Value:      "42",
					Properties: &jschema.RuleASTNodes{},
					Source:     jschema.RuleASTNodeSourceManual,
					Comment:    "foo",
				},
				{
					TokenType:  jschema.TokenTypeNumber,
					Value:      "3.14",
					Properties: &jschema.RuleASTNodes{},
					Source:     jschema.RuleASTNodeSourceManual,
				},
				{
					TokenType:  jschema.TokenTypeBoolean,
					Value:      "true",
					Properties: &jschema.RuleASTNodes{},
					Source:     jschema.RuleASTNodeSourceManual,
				},
				{
					TokenType:  jschema.TokenTypeNull,
					Value:      "null",
					Properties: &jschema.RuleASTNodes{},
					Source:     jschema.RuleASTNodeSourceManual,
					Comment:    "bar",
				},
				{
					TokenType:  jschema.TokenTypeShortcut,
					Value:      "@foo",
					Properties: &jschema.RuleASTNodes{},
					Source:     jschema.RuleASTNodeSourceManual,
				},
			},
			Source: jschema.RuleASTNodeSourceManual,
		}, e.ASTNode())
	})
}
