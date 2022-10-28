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
	actual := Enum{items: []EnumItem{
		NewEnumItem(bytes.Bytes(`"foo"`), ""),
		NewEnumItem(bytes.Bytes(`"bar"`), ""),
		NewEnumItem(bytes.Bytes(`"fizz"`), ""),
		NewEnumItem(bytes.Bytes(`"buzz"`), ""),
		NewEnumItem(bytes.Bytes(`123`), ""),
		NewEnumItem(bytes.Bytes(`45.67`), ""),
		NewEnumItem(bytes.Bytes(`true`), ""),
		NewEnumItem(bytes.Bytes(`false`), ""),
		NewEnumItem(bytes.Bytes(`null`), ""),
		NewEnumItem(bytes.Bytes(`"\""`), ""),
		NewEnumItem(bytes.Bytes(`"\\"`), ""),
		NewEnumItem(bytes.Bytes(`"\/"`), ""),
		NewEnumItem(bytes.Bytes(`"\b"`), ""),
		NewEnumItem(bytes.Bytes(`"\f"`), ""),
		NewEnumItem(bytes.Bytes(`"\n"`), ""),
		NewEnumItem(bytes.Bytes(`"\r"`), ""),
		NewEnumItem(bytes.Bytes(`"\t"`), ""),
		NewEnumItem(bytes.Bytes(`"\u0001"`), ""),
		NewEnumItem(bytes.Bytes(`"\u000A"`), ""),
		NewEnumItem(bytes.Bytes(`"\u20AC"`), ""),
		NewEnumItem(bytes.Bytes(`"\uD83C\uDFC6"`), ""),
	}}.
		String()

	assert.Equal(t, `enum: ["foo", "bar", "fizz", "buzz", 123, 45.67, true, false, null, "\"", "\\", "/", "\u0008", "\u000c", "\n", "\r", "\t", "\u0001", "\n", "€", "🏆"]`, actual)
}

func TestEnum_Append(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		c := NewEnum()
		assert.Equal(t, []EnumItem{}, c.items)

		c.Append(NewEnumItem(bytes.Bytes(`"foo"`), ""))
		assert.Equal(t, []EnumItem{
			NewEnumItem(bytes.Bytes(`"foo"`), ""),
		}, c.items)

		c.Append(NewEnumItem(bytes.Bytes(`"bar"`), ""))
		assert.Equal(t, []EnumItem{
			NewEnumItem(bytes.Bytes(`"foo"`), ""),
			NewEnumItem(bytes.Bytes(`"bar"`), ""),
		}, c.items)

		c.Append(NewEnumItem(bytes.Bytes(`"FoO"`), ""))
		assert.Equal(t, []EnumItem{
			NewEnumItem(bytes.Bytes(`"foo"`), ""),
			NewEnumItem(bytes.Bytes(`"bar"`), ""),
			NewEnumItem(bytes.Bytes(`"FoO"`), ""),
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
					c.Append(NewEnumItem(bytes.Bytes(`"foo"`), ""))
					c.Append(NewEnumItem(bytes.Bytes(given), ""))
				})
			})
		}
	})
}

func TestEnum_SetComment(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		e := &Enum{
			items: []EnumItem{
				NewEnumItem(bytes.Bytes(`"foo"`), ""),
				NewEnumItem(bytes.Bytes(`"bar"`), "old bar comment"),
			},
		}

		e.SetComment(0, "foo comment")
		e.SetComment(1, "new bar comment")

		assert.Equal(t, []EnumItem{
			NewEnumItem(bytes.Bytes(`"foo"`), "foo comment"),
			NewEnumItem(bytes.Bytes(`"bar"`), "new bar comment"),
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
				NewEnumItem(bytes.Bytes(`"foo"`), ""),
				NewEnumItem(bytes.Bytes(`"bar"`), ""),
			},
		}.
			Validate(bytes.Bytes(`"bar"`))
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithValue(t, errors.ErrDoesNotMatchAnyOfTheEnumValues, func() {
			Enum{
				items: []EnumItem{
					NewEnumItem(bytes.Bytes(`"foo"`), ""),
					NewEnumItem(bytes.Bytes(`"bar"`), ""),
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
			items: []EnumItem{
				NewEnumItem(bytes.Bytes(`"foo"`), ""),
				NewEnumItem(bytes.Bytes(`42`), "foo"),
				NewEnumItem(bytes.Bytes(`3.14`), ""),
				NewEnumItem(bytes.Bytes(`true`), ""),
				NewEnumItem(bytes.Bytes(`null`), "bar"),
				NewEnumItem(bytes.Bytes(`@foo`), ""),
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
