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
			JSONType:   jschema.JSONTypeArray,
			SchemaType: string(jschema.SchemaTypeEnum),
			Children: []jschema.ASTNode{
				{
					JSONType:   jschema.JSONTypeNull,
					SchemaType: string(jschema.SchemaTypeComment),
					Comment:    "first comment",
				},
				{
					JSONType:   jschema.JSONTypeString,
					SchemaType: string(jschema.SchemaTypeString),
					Value:      `"foo"`,
				},
				{
					JSONType:   jschema.JSONTypeNumber,
					SchemaType: string(jschema.SchemaTypeInteger),
					Value:      "42",
					Comment:    "42 comment",
				},
				{
					JSONType:   jschema.JSONTypeNumber,
					SchemaType: string(jschema.SchemaTypeFloat),
					Value:      "3.14",
				},
				{
					JSONType:   jschema.JSONTypeBoolean,
					SchemaType: string(jschema.SchemaTypeBoolean),
					Value:      "true",
				},
				{
					JSONType:   jschema.JSONTypeBoolean,
					SchemaType: string(jschema.SchemaTypeBoolean),
					Value:      "false",
					Comment:    "false comment",
				},
				{
					JSONType:   jschema.JSONTypeNull,
					SchemaType: string(jschema.SchemaTypeComment),
					Comment:    "before null comment",
				},
				{
					JSONType:   jschema.JSONTypeNull,
					SchemaType: string(jschema.SchemaTypeNull),
					Value:      "null",
					Comment:    "null comment",
				},
				{
					JSONType:   jschema.JSONTypeNull,
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
				{Comment: "Interline comment 1"},
				{Value: []byte(`"foo"`), Type: jschema.SchemaTypeString, Comment: "Foo comment"},
				{Value: []byte(`"bar"`), Type: jschema.SchemaTypeString, Comment: "Bar comment"},
				{Comment: "Interline comment 2"},
				{Value: []byte(`"fizz"`), Type: jschema.SchemaTypeString, Comment: "Fizz comment"},
				{Value: []byte(`"buzz"`), Type: jschema.SchemaTypeString, Comment: "Buzz comment"},
				{Comment: "Interline comment 3"},
			},
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				actual, err := New("", given).Values()

				require.NoError(t, err)
				assert.Equal(t, actual, expected)
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
