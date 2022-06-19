package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestNewEnum(t *testing.T) {
	c := NewEnum()
	assert.NotNil(t, c.items)
}

func TestEnum_IsJsonTypeCompatible(t *testing.T) {
	assert.True(t, Enum{}.IsJsonTypeCompatible(json.TypeString))
}

func TestEnum_Type(t *testing.T) {
	assert.Equal(t, EnumConstraintType, NewEnum().Type())
}

func TestEnum_String(t *testing.T) {
	actual := Enum{items: []enumItem{
		{value: bytes.Bytes(`"foo"`)},
		{value: bytes.Bytes(`"bar"`)},
		{value: bytes.Bytes(`"fizz"`)},
		{value: bytes.Bytes(`"buzz"`)},
	}}.
		String()

	assert.Equal(t, `enum: ["foo", "bar", "fizz", "buzz"]`, actual)
}

func TestEnum_Append(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		c := NewEnum()
		assert.Equal(t, []enumItem{}, c.items)

		c.Append(bytes.Bytes(`"foo"`))
		assert.Equal(t, []enumItem{
			{value: bytes.Bytes(`"foo"`)},
		}, c.items)

		c.Append(bytes.Bytes(`"bar"`))
		assert.Equal(t, []enumItem{
			{value: bytes.Bytes(`"foo"`)},
			{value: bytes.Bytes(`"bar"`)},
		}, c.items)

		c.Append(bytes.Bytes(`"FoO"`))
		assert.Equal(t, []enumItem{
			{value: bytes.Bytes(`"foo"`)},
			{value: bytes.Bytes(`"bar"`)},
			{value: bytes.Bytes(`"FoO"`)},
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
				{value: bytes.Bytes("foo")},
				{value: bytes.Bytes("bar"), comment: "old bar comment"},
			},
		}

		e.SetComment(0, "foo comment")
		e.SetComment(1, "new bar comment")

		assert.Equal(t, []enumItem{
			{value: bytes.Bytes("foo"), comment: "foo comment"},
			{value: bytes.Bytes("bar"), comment: "new bar comment"},
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
				{value: bytes.Bytes(`"foo"`)},
				{value: bytes.Bytes(`"bar"`)},
			},
		}.
			Validate(bytes.Bytes(`"bar"`))
	})

	t.Run("negative", func(t *testing.T) {
		assert.Panics(t, func() {
			Enum{
				items: []enumItem{
					{value: bytes.Bytes(`"foo"`)},
					{value: bytes.Bytes(`"bar"`)},
				},
			}.
				Validate(bytes.Bytes(`"fizz"`))
		})
	})
}

func TestEnum_ASTNode(t *testing.T) {
	e := Enum{
		items: []enumItem{
			{value: bytes.Bytes(`"foo"`)},
			{value: bytes.Bytes("42"), comment: "foo"},
			{value: bytes.Bytes("3.14")},
			{value: bytes.Bytes("true")},
			{value: bytes.Bytes("null"), comment: "bar"},
			{value: bytes.Bytes("@foo")},
		},
	}
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeArray,
		Properties: &jschema.RuleASTNodes{},
		Items: []jschema.RuleASTNode{
			{
				JSONType:   jschema.JSONTypeString,
				Value:      "foo",
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			},
			{
				JSONType:   jschema.JSONTypeNumber,
				Value:      "42",
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
				Comment:    "foo",
			},
			{
				JSONType:   jschema.JSONTypeNumber,
				Value:      "3.14",
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			},
			{
				JSONType:   jschema.JSONTypeBoolean,
				Value:      "true",
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			},
			{
				JSONType:   jschema.JSONTypeNull,
				Value:      "null",
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
				Comment:    "bar",
			},
			{
				JSONType:   jschema.JSONTypeShortcut,
				Value:      "@foo",
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			},
		},
		Source: jschema.RuleASTNodeSourceManual,
	}, e.ASTNode())
}
