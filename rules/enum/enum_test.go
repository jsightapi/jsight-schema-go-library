package enum

import (
	"testing"

	schema "github.com/jsightapi/jsight-schema-go-library"
	jbytes "github.com/jsightapi/jsight-schema-go-library/bytes"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnum_Len(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]uint{
			"[]":               2,
			"[]   \t  \n  \r ": 2,
			`[
				42,
				3.14,
				"foo",
				true,
				false,
				null
			]`: 65,
			"[42] something": 4,
			"":               0,
			`[
	// Interline comment 1
	1, // Comment for 1
	2, // Comment for 2

	// Interline comment 2
	3, // Comment for 3
	4  // Comment for 4
]`: 136,
			`[
		/* My
		   Pets */
		"CAT", /* My
		          Cat */
		"DOG", // Dog
		"PIG", // Pig

		// Wild animals
		"WOLF", // Wolf
		"LION", // Lion
		"TIGER" // Tiger
]`: 164,

			`["\u0061"]`: 10,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				actual, err := New("", given).Len()
				require.NoError(t, err)
				assert.Equal(t, expected, actual)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]string{
			`ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file 
	> 42
	--^`: "42",

			`ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file 
	> 42 [] foo
	--^`: "42 [] foo",

			`ERROR (code 303): Unexpected end of file
	in line 1 on file 
	> [
	--^`: "[",
		}

		for expected, given := range cc {
			t.Run(expected, func(t *testing.T) {
				_, err := New("", given).Len()
				assert.EqualError(t, err, expected)
			})
		}
	})
}

func TestEnum_Check(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		testList := []string{
			"[]",
			"[1]",
			"[1,2]",
			"[1,2,3]",
			"   [1,2,3]   ",
			"   [1,  2,  3]   ",
			"\n[1,2]",
			"[\n1,2]",
			"[1\n,2]",
			"[1,\n2]",
			"[1,2\n]",
			"[1,2]\n",
			`["aaa", "bbb", "ccc"]`,
			`[123, 45.67, "abc", true, false, null]`,
			`[
	123,
	45.67,
	"abc",
	true,
	false,
	null
]`,
			`[
	// Interline comment 1
	1, // Comment for 1
	2, // Comment for 2

	// Interline comment 2
	3, // Comment for 3
	4  // Comment for 4
]`,
			`[
		/* My
		   Pets */
		"CAT", /* My
		          Cat */
		"DOG", // Dog
		"PIG", // Pig

		// Wild animals
		"WOLF", // Wolf
		"LION", // Lion
		"TIGER" // Tiger
]`,
			`[3.14, 3.146]`,
			`["foo", "Foo"]`,
			`["a", "\u0062"]`,
			`["a", "\\u0061"]`,
		}

		for _, enum := range testList {
			t.Run(enum, func(t *testing.T) {
				err := New("enum", enum).Check()
				require.NoError(t, err)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]string{
			"123": `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file enum
	> 123
	--^`,
			`"abc"`: `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file enum
	> "abc"
	--^`,
			"true": `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file enum
	> true
	--^`,
			"false": `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file enum
	> false
	--^`,
			"null": `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file enum
	> null
	--^`,
			"{}": `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file enum
	> {}
	--^`,
			"[1,2,3] xxx": `ERROR (code 301): Invalid character "x" non-space byte after top-level value
	in line 1 on file enum
	> [1,2,3] xxx
	----------^`,
			"xxx [1,2,3]": `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file enum
	> xxx [1,2,3]
	--^`,
			"[1,]": `ERROR (code 301): Invalid character "]" looking for beginning of value
	in line 1 on file enum
	> [1,]
	-----^`,
			"[,1]": `ERROR (code 301): Invalid character "," looking for beginning of value
	in line 1 on file enum
	> [,1]
	---^`,
			"[ {} ]": `ERROR (code 301): Invalid character "{" looking for beginning of value
	in line 1 on file enum
	> [ {} ]
	----^`,
			"[ [] ]": `ERROR (code 301): Invalid character "[" looking for beginning of value
	in line 1 on file enum
	> [ [] ]
	----^`,

			"[1, 1]": `ERROR (code 810): 1 value duplicates in "enum"
	in line 1 on file enum
	> [1, 1]
	------^`,

			"[3.14, 3.14]": `ERROR (code 810): 3.14 value duplicates in "enum"
	in line 1 on file enum
	> [3.14, 3.14]
	---------^`,

			`["foo", "bar", "foo"]`: `ERROR (code 810): "foo" value duplicates in "enum"
	in line 1 on file enum
	> ["foo", "bar", "foo"]
	-----------------^`,

			"[true, true]": `ERROR (code 810): true value duplicates in "enum"
	in line 1 on file enum
	> [true, true]
	---------^`,

			"[null, null]": `ERROR (code 810): null value duplicates in "enum"
	in line 1 on file enum
	> [null, null]
	---------^`,

			"[   1\t,\n\n  1\t]": `ERROR (code 810): 1 value duplicates in "enum"
	in line 3 on file enum
	> 1	]
	--^`,

			`["a", "\u0061"]`: `ERROR (code 810): "\u0061" value duplicates in "enum"
	in line 1 on file enum
	> ["a", "\u0061"]
	--------^`,
		}

		for enum, expected := range cc {
			t.Run(enum, func(t *testing.T) {
				err := New("enum", enum).Check()
				assert.EqualError(t, err, expected)
			})
		}
	})
}

func TestEnum_GetAST(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		actual, err := New("", `[
	// first comment
	"foo",
	42, // 42 comment
	3.14,
	true,
	false, // false comment
	// before null comment
	null // null comment
	// last comment
]`).
			GetAST()

		require.NoError(t, err)
		assert.Equal(t, schema.ASTNode{
			TokenType:  schema.TokenTypeArray,
			SchemaType: string(schema.SchemaTypeEnum),
			Children: []schema.ASTNode{
				{
					TokenType:  schema.TokenTypeNull,
					SchemaType: string(schema.SchemaTypeComment),
					Comment:    "first comment",
				},
				{
					TokenType:  schema.TokenTypeString,
					SchemaType: string(schema.SchemaTypeString),
					Value:      `"foo"`,
				},
				{
					TokenType:  schema.TokenTypeNumber,
					SchemaType: string(schema.SchemaTypeInteger),
					Value:      "42",
					Comment:    "42 comment",
				},
				{
					TokenType:  schema.TokenTypeNumber,
					SchemaType: string(schema.SchemaTypeFloat),
					Value:      "3.14",
				},
				{
					TokenType:  schema.TokenTypeBoolean,
					SchemaType: string(schema.SchemaTypeBoolean),
					Value:      "true",
				},
				{
					TokenType:  schema.TokenTypeBoolean,
					SchemaType: string(schema.SchemaTypeBoolean),
					Value:      "false",
					Comment:    "false comment",
				},
				{
					TokenType:  schema.TokenTypeNull,
					SchemaType: string(schema.SchemaTypeComment),
					Comment:    "before null comment",
				},
				{
					TokenType:  schema.TokenTypeNull,
					SchemaType: string(schema.SchemaTypeNull),
					Value:      "null",
					Comment:    "null comment",
				},
				{
					TokenType:  schema.TokenTypeNull,
					SchemaType: string(schema.SchemaTypeComment),
					Comment:    "last comment",
				},
			},
		}, actual)
	})
}

func TestEnum_Values(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string][]Value{
			`[
	"foo",
	42,
	3.14,
	true,
	false,
	null
]`: {
				{Value: jbytes.NewBytes(`"foo"`), Type: schema.SchemaTypeString},
				{Value: jbytes.NewBytes("42"), Type: schema.SchemaTypeInteger},
				{Value: jbytes.NewBytes("3.14"), Type: schema.SchemaTypeFloat},
				{Value: jbytes.NewBytes("true"), Type: schema.SchemaTypeBoolean},
				{Value: jbytes.NewBytes("false"), Type: schema.SchemaTypeBoolean},
				{Value: jbytes.NewBytes("null"), Type: schema.SchemaTypeNull},
			},

			`[
	// Interline comment 1
	"foo", // Foo comment
	"bar", // Bar comment

	// Interline comment 2
	"fizz", // Fizz comment
	"buzz"  // Buzz comment

	// Interline comment 3
]`: {
				{Comment: "Interline comment 1", Type: schema.SchemaTypeComment},
				{Value: jbytes.NewBytes(`"foo"`), Type: schema.SchemaTypeString, Comment: "Foo comment"},
				{Value: jbytes.NewBytes(`"bar"`), Type: schema.SchemaTypeString, Comment: "Bar comment"},
				{Comment: "Interline comment 2", Type: schema.SchemaTypeComment},
				{Value: jbytes.NewBytes(`"fizz"`), Type: schema.SchemaTypeString, Comment: "Fizz comment"},
				{Value: jbytes.NewBytes(`"buzz"`), Type: schema.SchemaTypeString, Comment: "Buzz comment"},
				{Comment: "Interline comment 3", Type: schema.SchemaTypeComment},
			},

			`[
		/* My
		   Pets */
		"CAT", /* My
		          Cat */
		"DOG", // Dog
		"PIG", // Pig

		// Wild animals
		"WOLF", // Wolf
		"LION", // Lion
		"TIGER" // Tiger
]`: {
				{Comment: "My\n\t\t   Pets", Type: schema.SchemaTypeComment},
				{Value: jbytes.NewBytes(`"CAT"`), Type: schema.SchemaTypeString, Comment: "My\n\t\t          Cat"},
				{Value: jbytes.NewBytes(`"DOG"`), Type: schema.SchemaTypeString, Comment: "Dog"},
				{Value: jbytes.NewBytes(`"PIG"`), Type: schema.SchemaTypeString, Comment: "Pig"},
				{Comment: "Wild animals", Type: schema.SchemaTypeComment},
				{Value: jbytes.NewBytes(`"WOLF"`), Type: schema.SchemaTypeString, Comment: "Wolf"},
				{Value: jbytes.NewBytes(`"LION"`), Type: schema.SchemaTypeString, Comment: "Lion"},
				{Value: jbytes.NewBytes(`"TIGER"`), Type: schema.SchemaTypeString, Comment: "Tiger"},
			},
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				actual, err := New("", given).Values()

				require.NoError(t, err)
				assert.Equal(t, expected, actual)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		_, err := New("", "123").Values()
		assert.EqualError(t, err, `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file 
	> 123
	--^`)
	})
}
