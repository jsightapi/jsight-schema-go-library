package enum

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jschema "github.com/jsightapi/jsight-schema-go-library"
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
		assert.Equal(t, jschema.ASTNode{
			TokenType:  jschema.TokenTypeArray,
			SchemaType: string(jschema.SchemaTypeEnum),
			Children: []jschema.ASTNode{
				{
					TokenType:  jschema.TokenTypeNull,
					SchemaType: string(jschema.SchemaTypeComment),
					Comment:    "first comment",
				},
				{
					TokenType:  jschema.TokenTypeString,
					SchemaType: string(jschema.SchemaTypeString),
					Value:      `"foo"`,
				},
				{
					TokenType:  jschema.TokenTypeNumber,
					SchemaType: string(jschema.SchemaTypeInteger),
					Value:      "42",
					Comment:    "42 comment",
				},
				{
					TokenType:  jschema.TokenTypeNumber,
					SchemaType: string(jschema.SchemaTypeFloat),
					Value:      "3.14",
				},
				{
					TokenType:  jschema.TokenTypeBoolean,
					SchemaType: string(jschema.SchemaTypeBoolean),
					Value:      "true",
				},
				{
					TokenType:  jschema.TokenTypeBoolean,
					SchemaType: string(jschema.SchemaTypeBoolean),
					Value:      "false",
					Comment:    "false comment",
				},
				{
					TokenType:  jschema.TokenTypeNull,
					SchemaType: string(jschema.SchemaTypeComment),
					Comment:    "before null comment",
				},
				{
					TokenType:  jschema.TokenTypeNull,
					SchemaType: string(jschema.SchemaTypeNull),
					Value:      "null",
					Comment:    "null comment",
				},
				{
					TokenType:  jschema.TokenTypeNull,
					SchemaType: string(jschema.SchemaTypeComment),
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
				{Value: []byte(`"foo"`), Type: jschema.SchemaTypeString},
				{Value: []byte("42"), Type: jschema.SchemaTypeInteger},
				{Value: []byte("3.14"), Type: jschema.SchemaTypeFloat},
				{Value: []byte("true"), Type: jschema.SchemaTypeBoolean},
				{Value: []byte("false"), Type: jschema.SchemaTypeBoolean},
				{Value: []byte("null"), Type: jschema.SchemaTypeNull},
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
				{Comment: "Interline comment 1", Type: jschema.SchemaTypeComment},
				{Value: []byte(`"foo"`), Type: jschema.SchemaTypeString, Comment: "Foo comment"},
				{Value: []byte(`"bar"`), Type: jschema.SchemaTypeString, Comment: "Bar comment"},
				{Comment: "Interline comment 2", Type: jschema.SchemaTypeComment},
				{Value: []byte(`"fizz"`), Type: jschema.SchemaTypeString, Comment: "Fizz comment"},
				{Value: []byte(`"buzz"`), Type: jschema.SchemaTypeString, Comment: "Buzz comment"},
				{Comment: "Interline comment 3", Type: jschema.SchemaTypeComment},
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
				{Comment: "My\n\t\t   Pets", Type: jschema.SchemaTypeComment},
				{Value: []byte(`"CAT"`), Type: jschema.SchemaTypeString, Comment: "My\n\t\t          Cat"},
				{Value: []byte(`"DOG"`), Type: jschema.SchemaTypeString, Comment: "Dog"},
				{Value: []byte(`"PIG"`), Type: jschema.SchemaTypeString, Comment: "Pig"},
				{Comment: "Wild animals", Type: jschema.SchemaTypeComment},
				{Value: []byte(`"WOLF"`), Type: jschema.SchemaTypeString, Comment: "Wolf"},
				{Value: []byte(`"LION"`), Type: jschema.SchemaTypeString, Comment: "Lion"},
				{Value: []byte(`"TIGER"`), Type: jschema.SchemaTypeString, Comment: "Tiger"},
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
