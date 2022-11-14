package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/json"

	schema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
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
	actual := Enum{items: []EnumItem{
		NewEnumItem(bytes.NewBytes(`"foo"`), ""),
		NewEnumItem(bytes.NewBytes(`"bar"`), ""),
		NewEnumItem(bytes.NewBytes(`"fizz"`), ""),
		NewEnumItem(bytes.NewBytes(`"buzz"`), ""),
		NewEnumItem(bytes.NewBytes(`123`), ""),
		NewEnumItem(bytes.NewBytes(`45.67`), ""),
		NewEnumItem(bytes.NewBytes(`true`), ""),
		NewEnumItem(bytes.NewBytes(`false`), ""),
		NewEnumItem(bytes.NewBytes(`null`), ""),
		NewEnumItem(bytes.NewBytes(`"\""`), ""),
		NewEnumItem(bytes.NewBytes(`"\\"`), ""),
		NewEnumItem(bytes.NewBytes(`"\/"`), ""),
		NewEnumItem(bytes.NewBytes(`"\b"`), ""),
		NewEnumItem(bytes.NewBytes(`"\f"`), ""),
		NewEnumItem(bytes.NewBytes(`"\n"`), ""),
		NewEnumItem(bytes.NewBytes(`"\r"`), ""),
		NewEnumItem(bytes.NewBytes(`"\t"`), ""),
		NewEnumItem(bytes.NewBytes(`"\u0001"`), ""),
		NewEnumItem(bytes.NewBytes(`"\u000A"`), ""),
		NewEnumItem(bytes.NewBytes(`"\u20AC"`), ""),
		NewEnumItem(bytes.NewBytes(`"\uD83C\uDFC6"`), ""),
	}}.
		String()

	assert.Equal(t, `enum: ["foo", "bar", "fizz", "buzz", 123, 45.67, true, false, null, "\"", "\\", "/", "\u0008", "\u000c", "\n", "\r", "\t", "\u0001", "\n", "€", "🏆"]`, actual)
}

func TestEnum_Append(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		c := NewEnum()
		assert.Equal(t, []EnumItem{}, c.items)

		c.Append(NewEnumItem(bytes.NewBytes(`"foo"`), ""))
		assert.Equal(t, []EnumItem{
			NewEnumItem(bytes.NewBytes(`"foo"`), ""),
		}, c.items)

		c.Append(NewEnumItem(bytes.NewBytes(`"bar"`), ""))
		assert.Equal(t, []EnumItem{
			NewEnumItem(bytes.NewBytes(`"foo"`), ""),
			NewEnumItem(bytes.NewBytes(`"bar"`), ""),
		}, c.items)

		c.Append(NewEnumItem(bytes.NewBytes(`"FoO"`), ""))
		assert.Equal(t, []EnumItem{
			NewEnumItem(bytes.NewBytes(`"foo"`), ""),
			NewEnumItem(bytes.NewBytes(`"bar"`), ""),
			NewEnumItem(bytes.NewBytes(`"FoO"`), ""),
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
					c.Append(NewEnumItem(bytes.NewBytes(`"foo"`), ""))
					c.Append(NewEnumItem(bytes.NewBytes(given), ""))
				})
			})
		}
	})
}

func TestEnum_SetComment(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		e := &Enum{
			items: []EnumItem{
				NewEnumItem(bytes.NewBytes(`"foo"`), ""),
				NewEnumItem(bytes.NewBytes(`"bar"`), "old bar comment"),
			},
		}

		e.SetComment(0, "foo comment")
		e.SetComment(1, "new bar comment")

		assert.Equal(t, []EnumItem{
			NewEnumItem(bytes.NewBytes(`"foo"`), "foo comment"),
			NewEnumItem(bytes.NewBytes(`"bar"`), "new bar comment"),
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
			items: []EnumItem{
				NewEnumItem(bytes.NewBytes(`"foo"`), ""),
				NewEnumItem(bytes.NewBytes(`"bar"`), ""),
			},
		}.
			Validate(bytes.NewBytes(`"bar"`))
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithValue(t, errors.ErrDoesNotMatchAnyOfTheEnumValues, func() {
			Enum{
				items: []EnumItem{
					NewEnumItem(bytes.NewBytes(`"foo"`), ""),
					NewEnumItem(bytes.NewBytes(`"bar"`), ""),
				},
			}.
				Validate(bytes.NewBytes(`"fizz"`))
		})
	})
}

func TestEnum_ASTNode(t *testing.T) {
	t.Run("with rule name", func(t *testing.T) {
		e := Enum{
			ruleName: "@foo",
		}

		assert.Equal(t, schema.RuleASTNode{
			TokenType:  schema.TokenTypeShortcut,
			Value:      "@foo",
			Properties: &schema.RuleASTNodes{},
			Source:     schema.RuleASTNodeSourceManual,
		}, e.ASTNode())
	})

	t.Run("without rule name", func(t *testing.T) {
		e := Enum{
			items: []EnumItem{
				NewEnumItem(bytes.NewBytes(`"foo"`), ""),
				NewEnumItem(bytes.NewBytes(`42`), "foo"),
				NewEnumItem(bytes.NewBytes(`3.14`), ""),
				NewEnumItem(bytes.NewBytes(`true`), ""),
				NewEnumItem(bytes.NewBytes(`null`), "bar"),
				NewEnumItem(bytes.NewBytes(`@foo`), ""),
			},
		}
		assert.Equal(t, schema.RuleASTNode{
			TokenType:  schema.TokenTypeArray,
			Properties: &schema.RuleASTNodes{},
			Items: []schema.RuleASTNode{
				{
					TokenType:  schema.TokenTypeString,
					Value:      "foo",
					Properties: &schema.RuleASTNodes{},
					Source:     schema.RuleASTNodeSourceManual,
				},
				{
					TokenType:  schema.TokenTypeNumber,
					Value:      "42",
					Properties: &schema.RuleASTNodes{},
					Source:     schema.RuleASTNodeSourceManual,
					Comment:    "foo",
				},
				{
					TokenType:  schema.TokenTypeNumber,
					Value:      "3.14",
					Properties: &schema.RuleASTNodes{},
					Source:     schema.RuleASTNodeSourceManual,
				},
				{
					TokenType:  schema.TokenTypeBoolean,
					Value:      "true",
					Properties: &schema.RuleASTNodes{},
					Source:     schema.RuleASTNodeSourceManual,
				},
				{
					TokenType:  schema.TokenTypeNull,
					Value:      "null",
					Properties: &schema.RuleASTNodes{},
					Source:     schema.RuleASTNodeSourceManual,
					Comment:    "bar",
				},
				{
					TokenType:  schema.TokenTypeShortcut,
					Value:      "@foo",
					Properties: &schema.RuleASTNodes{},
					Source:     schema.RuleASTNodeSourceManual,
				},
			},
			Source: schema.RuleASTNodeSourceManual,
		}, e.ASTNode())
	})
}
