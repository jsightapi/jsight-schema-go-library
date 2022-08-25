package jschema

import (
	stdErrors "errors"
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/formats/json"
	"github.com/jsightapi/jsight-schema-go-library/internal/mocks"
	schemaMocks "github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/mocks"
	internalSchema "github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema"
	"github.com/jsightapi/jsight-schema-go-library/notations/regex"
	"github.com/jsightapi/jsight-schema-go-library/rules/enum"
)

func ExampleSchema() {
	s := MustNew("root", `{"foo": @Fizz,"bar": @Buzz}`)

	l, err := s.Len()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	fmt.Println(l)

	err = s.AddType("@Fizz", MustNew("fizz", `{"fizz": 1}`))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	err = s.AddType("@Buzz", MustNew("buzz", `{"buzz": 2}`))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	err = s.Check()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	err = s.Validate(json.MustNew("json", `{"foo":{"fizz":42},"bar":{"buzz":42}}`))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	// Output: 27
}

func TestSchema_Len(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]uint{
			`
{
	"key": 123 // {min: 1}
}
some extra text
`: 28,
			`@pig // {or: ["@dog", "@pig"]}`:  30,
			`@pig, // {or: ["@dog", "@pig"]}`: 4,
			`@pig, // {or: ["@dog", "@pig"]}
some extra text`: 4,
			`42 /*
	{nullable: true}
*/
some extra text`: 26,
			"[]  // {minItems: 0} - Description":                                  34,
			"[]  // {minItems: 0} - Description ":                                 34,
			"[]  // {minItems: 0} - Description  ":                                34,
			"[]  // {minItems: 0} - Description \n some data":                     34,
			`"userType2": 12 // {type: "@catId", optional: true, nullable: true}`: 11,
			`[
	{} // {type: @json}
]`: 24,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				l, err := MustNew("foo", given).Len()
				require.NoError(t, err)
				assert.Equal(t, int(expected), int(l))
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		_, err := MustNew("foo", "invalid").Len()
		assert.EqualError(t, err, `ERROR (code 301): Invalid character "i" looking for beginning of value
	in line 1 on file foo
	> invalid
	--^`)
	})
}

func BenchmarkSchema_Len(b *testing.B) {
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		s := MustNew("foo", `[
  {
    "id": 1,
    "first_name": "Cecilia",
    "last_name": "Maudson",
    "email": "cmaudson0@dedecms.com",
    "gender": "Female",
    "ip_address": "14.224.72.249"
  }
]`)
		b.StartTimer()
		l, err := s.Len()
		require.NoError(b, err)
		assert.Equal(b, 177, int(l))
	}
}

func TestSchema_Example(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]struct {
			enums    map[string]string
			types    map[string]string
			expected string
		}{
			`
{ //{allOf: ["@allOf1", "@allOf2"]}
	"i": 123, // {min: 1}
	"s": "str",
	"b": true,
	"n": null,
	"a": [1, "str", false, null],
	"o": {
		"ii": 999 // {max: 999}
	},
	"or_full": "foo", // {or: [{"type": "string"}, {"type": "integer"}]}
	"or_short": "foo", // {or: ["string", "integer"]}
	"shortcut": @foo,
	"shortcut_or": @foo | @bar,
	"enum": 1, // {enum: [1, 2, 3]}
	"enum_rule": 2, // {enum: @enum}
	"recursion": @recursion,
	"deep_recursion": @deep_recursion,
	@keyName: 100500,
	"@keyName": "should not change"
}
`: {
				enums: map[string]string{
					"@enum": "[1, 2, 3]",
				},
				types: map[string]string{
					"@foo": `{
	"foo": 42
}`,
					"@bar": `{
	"bar": 42
}`,
					"@recursion": `{
	"recursion": @recursion // {optional: true}
}`,
					"@deep_recursion": `{
	"bar": @nested
}`,
					"@nested": `{
	"fizz": @deep_recursion
}`,
					"@keyName": `"key_name_1" // {regex: "key_name_\\d+"}`,
					"@allOf1": `{
	"allOf1": 42
}`,
					"@allOf2": `{
	"allOf2": @recursion // {optional: true}
}`,
				},
				expected: `{
	"i": 123,
	"s": "str",
	"b": true,
	"n": null,
	"a": [
		1,
		"str",
		false,
		null
	],
	"o": {
		"ii": 999
	},
	"or_full": "foo",
	"or_short": "foo",
	"shortcut": {
		"foo": 42
	},
	"shortcut_or": {
		"foo": 42
	},
	"enum": 1,
	"enum_rule": 2,
	"recursion": {
		"recursion": {}
	},
	"deep_recursion": {
		"bar": {
			"fizz": {
				"bar": {}
			}
		}
	},
	"key_name_1": 100500,
	"@keyName": "should not change",
	"allOf1": 42,
	"allOf2": {
		"recursion": {}
	}
}`,
			},

			`{
	"main1": @main, // {optional: true}
	"main2": @main // {optional: true}
}`: {
				expected: `{
	"main1": {
		"main1": {},
		"main2": {}
	},
	"main2": {
		"main1": {},
		"main2": {}
	}
}`,
			},

			`"\" \\ /"`: {
				expected: `"\" \\ /"`,
			},
		}

		for given, c := range cc {
			t.Run(given, func(t *testing.T) {
				s := MustNew("@main", given)

				for n, b := range c.enums {
					require.NoError(t, s.AddRule(n, enum.MustNew(n, b)))
				}

				for n, b := range c.types {
					require.NoError(t, s.AddType(n, MustNew(n, b)))
				}
				require.NoError(t, s.AddType("@main", s))

				actual, err := s.Example()
				require.NoError(t, err)
				assert.JSONEq(t, c.expected, string(actual))
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		_, err := MustNew("schema", "invalid").Example()
		assert.EqualError(t, err, `ERROR (code 301): Invalid character "i" looking for beginning of value
	in line 1 on file schema
	> invalid
	--^`)
	})
}

func TestSchema_AddType(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		t.Run("jschema", func(t *testing.T) {
			root := MustNew("", `{"foo": @foo}`)
			typ := MustNew("", "123")
			err := root.AddType("@foo", typ)
			require.NoError(t, err)

			require.NotNil(t, root.inner)
			actualType, err := root.inner.Type("@foo")
			require.NoError(t, err)
			assert.Equal(t, typ.inner, actualType)
		})

		t.Run("regex", func(t *testing.T) {
			root := MustNew("", `{"foo": @foo}`)
			typ := regex.MustNew("", "/foo-\\d/")
			err := root.AddType("@foo", typ)
			require.NoError(t, err)

			require.NotNil(t, root.inner)
		})
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("invalid schema", func(t *testing.T) {
			err := MustNew("", "42").AddType("invalid", nil)
			assert.EqualError(t, err, "schema should be JSight or Regex schema, but <nil> given")
		})

		t.Run("invalid schema name", func(t *testing.T) {
			err := MustNew("", "42").AddType("invalid", MustNew("invalid", "42"))
			assert.EqualError(t, err, "Invalid schema name (invalid)")
		})
	})
}

func TestSchema_AddRule(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		const name = "foo"
		r := mocks.NewRule(t)
		r.On("Check").Return(nil)
		s := MustNew("", "content")

		err := s.AddRule(name, r)

		require.NoError(t, err)
		assert.Len(t, s.rules, 1)
		assert.Contains(t, s.rules, name)
		assert.Same(t, r, s.rules[name])
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("already compiled", func(t *testing.T) {
			s := MustNew("foo", "content")
			s.inner = &internalSchema.Schema{}

			err := s.AddRule("foo", mocks.NewRule(t))

			assert.EqualError(t, err, "schema is already compiled")
			assert.Len(t, s.rules, 0)
		})

		t.Run("nil rule", func(t *testing.T) {
			s := MustNew("", "content")

			err := s.AddRule("", nil)

			assert.EqualError(t, err, "rule is nil")
			assert.Len(t, s.rules, 0)
		})

		t.Run("invalid rule", func(t *testing.T) {
			r := mocks.NewRule(t)
			r.On("Check").Return(stdErrors.New("fake error"))
			s := MustNew("", "content")

			err := s.AddRule("", r)

			assert.EqualError(t, err, "fake error")
			assert.Len(t, s.rules, 0)
		})
	})
}

//goland:noinspection HttpUrlsUsage
func TestSchema_Check(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]struct {
			types map[string]string
			enums map[string]string
		}{
			`{"foo": "bar"}`:         {},
			`{} // {type: "object"}`: {},
			`[] // {type: "array"}`:  {},
			"@foo": {
				types: map[string]string{
					"@foo": `{"foo": "bar"}`,
				},
			},
			`{} // {or: [{type: "object"}, {type: "array"}]}`:     {},
			`[] // {or: [{type: "object"}, {type: "array"}]}`:     {},
			`{} // {or: [{type: "object"}, {type: "string"}]}`:    {},
			`"foo" // {or: [{type: "object"}, {type: "string"}]}`: {},
			`[] // {or: [{type: "array"}, {type: "string"}]}`:     {},
			`"foo" // {or: [{type: "array"}, {type: "string"}]}`:  {},
			`"CAT-123" // {type: "@catId"}`: {
				types: map[string]string{
					"@catId": `"CAT-123"`,
				},
			},
			`"foo" // {or: [{type: "string"}, {type: "@foo"}]}`: {
				types: map[string]string{
					"@foo": `{"key": "value"}`,
				},
			},
			"@foo | @bar": {
				types: map[string]string{
					"@foo": `{"foo": "bar"}`,
					"@bar": `{"foo": "bar"}`,
				},
			},
			`{"myCat": @cat}`: {
				types: map[string]string{
					"@cat": `{"foo": "bar"}`,
				},
			},
			`{
				"myCatList": [
					@cat
				]
			}`: {
				types: map[string]string{
					"@cat": `{"foo": "bar"}`,
				},
			},
			`{
				"myCat": @cat // {optional: true}
			}`: {
				types: map[string]string{
					"@cat": "42",
				},
			},
			`[
				@cat | @dog | @frog
			]`: {
				types: map[string]string{
					"@cat":  `{"foo": "bar"}`,
					"@dog":  `{"foo": "bar"}`,
					"@frog": `{"foo": "bar"}`,
				},
			},
			`{
				"myPet": @cat | @dog // {optional: true}
			}`: {
				types: map[string]string{
					"@cat": `{"foo": "bar"}`,
					"@dog": `{"foo": "bar"}`,
				},
			},
			`{
				"myPetId": "CAT-123" // {or: ["@catId", "@dogId"]}
			}`: {
				types: map[string]string{
					"@catId": `"CAT-123"`,
					"@dogId": `"DOG-123"`,
				},
			},
			`{
				"@catsEmail" : @cat
			}`: {
				types: map[string]string{
					"@cat": `{"foo": "bar"}`,
				},
			},
			`{
				@catsEmail : @cat
			}`: {
				types: map[string]string{
					"@cat":       `{"foo": "bar"}`,
					"@catsEmail": `"email@address.com"`,
				},
			},
			"42 // {const: true}":  {},
			"{} // {const: false}": {},
			`{ // {const: false}
				"foo": "bar"
			}`: {},
			"[] // {const: false}": {},
			`[ // {const: false}
				1,
				2,
				3
			]`: {},
			`42 // {type: "@foo", const: false}`: {
				types: map[string]string{
					"@foo": "42",
				},
			},
			"@foo // {const: false}": {
				types: map[string]string{
					"@foo": `{"foo": "bar"}`,
				},
			},
			"@foo | @bar // {const: false}": {
				types: map[string]string{
					"@foo": `{"foo": "bar"}`,
					"@bar": `{"foo": "bar"}`,
				},
			},
			`{
				"data": "abc" /* {
					or: [
						{type: "string" , maxLength: 3},
						{type: "integer", min: 0}
					]
				} */
			}`: {},
			`[ // {type: "array", maxItems: 100}
		1, // {type: "mixed", or: [{type: "integer"}, {type: "string"}]}
		2 // {or: [{type: "integer"}, {type: "string"}]}
]`: {
				types: map[string]string{
					"@dog": `{"foo": "bar"}`,
					"@pig": `{"foo": "bar"}`,
				},
			},
			`[ // {type: "array", maxItems: 100}
		@dog | @pig
]`: {
				types: map[string]string{
					"@dog": `{"foo": "bar"}`,
					"@pig": `{"foo": "bar"}`,
				},
			},
			`{
	"tags": [
		"@cats"
	],
	"query"  : @query,
	"request": @httpRequest
}`: {
				types: map[string]string{
					"@query":       `{"foo": "bar"}`,
					"@httpRequest": `{"foo": "bar"}`,
				},
			},

			`"2021-01-08" // {type: "date"}`: {},
			`[
	"2021-01-08" // {type: "date"}
]`: {},
			`{
	"foo": "2021-01-08" // {type: "date"}
}`: {},

			`"2021-01-08T12:50:45+06:00" // {type: "datetime"}`: {},
			`[
	"2021-01-08T12:50:45+06:00" // {type: "datetime"}
]`: {},
			`{
	"foo": "2021-01-08T12:50:45+06:00" // {type: "datetime"}
}`: {},

			`{
  "id1": 1, // {type: "@id", nullable: true}
  "id2": @id, // {nullable: true}
  "id3": @id1 | @id2, // {nullable: true}
  "size": 1, // {enum: [1,2,3], nullable: true}
  "choice": 1 // {or: [{type: "integer"}, {type: "string"}]}
}`: {
				types: map[string]string{
					"@id":  "123",
					"@id1": "[]",
					"@id2": "{}",
				},
			},
			`42 // {type: "any", nullable: true}`: {},
			`{
	"foo": 123 /* {or: [
		{min: 100},
		{type: "string"}
	]} */
}`: {},
			`42 // {or: ["integer", "string"]}`: {},
			"@bar": {
				types: map[string]string{
					"@bar": `42 // {or: ["integer", "string"]}`,
				},
			},
			"1 // {enum : [1]}": {},
			`{
	"foo": 2 // {nullable: false, optional: true}
}`: {},
			`"5" // {enum: ["5", 5]}`: {},
			`{ // {allOf: "@bar"}
	"foo": 1
}`: {
				types: map[string]string{
					"@bar": `{ // {allOf: "@fizz"}
	"bar": 2 // {or: ["integer", "string"]}
}`,
					"@fizz": `{
	"fizz": 3 // {or: ["integer", "string"]}
}`,
				},
			},

			`"foo" // {enum: @enum}`: {
				enums: map[string]string{
					"@enum": `["foo", "bar"]`,
				},
			},

			`{
	"foo": "foo" // {enum: @enum}
}`: {
				enums: map[string]string{
					"@enum": `["foo", "bar"]`,
				},
			},

			`3.14 // {type: "decimal", precision: 2}`: {},

			// Valid recursions.
			`{
	"foo": @bar
}`: {
				types: map[string]string{
					"@bar": `{
	"bar": @main // {optional: true}
}`,
				},
			},

			`{
	"foo": [@main]
}`: {},

			`{
	"foo": @fizz | @buzz
}`: {
				types: map[string]string{
					"@fizz": `{
	"fizz": @main
}`,
					"@buzz": `{
	"buzz": 42
}`,
				},
			},

			`1 /* {or: [
	{type: "string"},
	{enum: [1,2,3]}
]} */`: {},

			`"foo" /* {or: [
	{type: "string"},
	{enum: [1,2,3]}
]} */`: {},

			`1 /* {or: [
	{type: "string"},
	{enum: @enum}
]} */`: {
				enums: map[string]string{
					"@enum": "[1, 2, 3]",
				},
			},

			`"foo" /* {or: [
	{type: "string"},
	{enum: @enum}
]} */`: {
				enums: map[string]string{
					"@enum": "[1, 2, 3]",
				},
			},

			`"foo" /* {or: [
	{type: "string"},
	{enum: @enum}
]} - comment */`: {
				enums: map[string]string{
					"@enum": "[1, 2, 3]",
				},
			},

			`"foo" /* {or: [
	{type: "string"},
	{enum: @enum}
]} - multi
	line
	comment */`: {
				enums: map[string]string{
					"@enum": "[1, 2, 3]",
				},
			},

			`{
  @catId: 1,
  "@catId": 1
}`: {
				types: map[string]string{
					"@catId": `"foo"`,
				},
			},

			"\"a\" // {enum: [\"a\", \"\u0062\"]}": {},
		}

		for content, c := range cc {
			t.Run(content, func(t *testing.T) {
				s := MustNew("@main", content)

				for n, c := range c.enums {
					require.NoError(t, s.AddRule(n, enum.MustNew(n, c)))
				}

				for n, c := range c.types {
					require.NoError(t, s.AddType(n, MustNew(n, c)))
				}
				require.NoError(t, s.AddType("@main", s))

				require.NoError(t, s.Check())
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]struct {
			types map[string]string
			rules map[string]string
			given string
		}{
			`ERROR (code 301): Invalid character "i" looking for beginning of value
	in line 1 on file 
	> invalid
	--^`: {
				given: "invalid",
			},
			`ERROR (code 1302): Type "@int" not found
	in line 2 on file 
	> "aaa": 111 // {type: "@int"}
	---------^`: {
				given: `{
		"aaa": 111 // {type: "@int"}
	}`,
			},

			`ERROR (code 1302): Type "@foo" not found
	in line 1 on file 
	> @foo
	--^`: {
				given: "@foo",
			},

			`ERROR (code 1302): Type "@foo" not found
	in line 1 on file 
	> 42 // {type: "@foo"}
	--^`: {
				given: `42 // {type: "@foo"}`,
			},

			`ERROR (code 804): You cannot place a RULE on lines that contain more than one EXAMPLE node to which any RULES can apply. The only exception is when an object key and its value are found in one line.
	in line 1 on file 
	> {"foo": "bar"} // {const: false}
	----------------------------^`: {
				given: `{"foo": "bar"} // {const: false}`,
			},

			`ERROR (code 304): Annotation not allowed here
	in line 1 on file 
	> [1, 2, 3] // {const: false}
	------------^`: {
				given: "[1, 2, 3] // {const: false}",
			},

			`ERROR (code 1117): The "const" constraint can't be used for the "object" type
	in line 1 on file 
	> {} // {const: true}
	--^`: {
				given: "{} // {const: true}",
			},

			`ERROR (code 1117): The "const" constraint can't be used for the "object" type
	in line 1 on file 
	> { // {const: true}
	--^`: {
				given: `{ // {const: true}
	"foo": "bar"
}`,
			},

			`ERROR (code 1117): The "const" constraint can't be used for the "array" type
	in line 1 on file 
	> [] // {const: true}
	--^`: {
				given: "[] // {const: true}",
			},

			`ERROR (code 1117): The "const" constraint can't be used for the "array" type
	in line 1 on file 
	> [ // {const: true}
	--^`: {
				given: `[ // {const: true}
	1,
	2,
	3
]`,
			},

			`ERROR (code 1102): Invalid rule set shared with a type reference
	in line 1 on file 
	> @foo // {const: true}
	--^`: {
				given: "@foo // {const: true}",
			},

			`ERROR (code 1103): Invalid rule set shared with "or"
	in line 1 on file 
	> @foo | @bar // {const: true}
	--^`: {
				given: "@foo | @bar // {const: true}",
			},

			`ERROR (code 1114): Not found the rule "or" for the "mixed" type
	in line 1 on file 
	> 42 // {type: "mixed", const: true}
	--^`: {
				given: `42 // {type: "mixed", const: true}`,
			},

			`ERROR (code 1103): Invalid rule set shared with "or"
	in line 1 on file 
	> 42 // {type: "mixed", or: ["@foo", "@bar"], const: true}
	--^`: {
				given: `42 // {type: "mixed", or: ["@foo", "@bar"], const: true}`,
			},

			`ERROR (code 1103): Invalid rule set shared with "or"
	in line 1 on file 
	> 42 // {or: [{type: "integer"}, {type: "string"}], const: true}
	--^`: {
				given: `42 // {or: [{type: "integer"}, {type: "string"}], const: true}`,
			},

			`ERROR (code 1102): Invalid rule set shared with a type reference
	in line 1 on file 
	> 42 // {type: "@foo", const: true}
	--^`: {
				given: `42 // {type: "@foo", const: true}`,
			},

			`ERROR (code 301): Invalid character "/" looking for beginning of string
	in line 3 on file 
	> // inline comment
	--^`: {
				given: `{
	"foo": "bar",
	// inline comment
	"fizz": "buzz"
}`,
			},

			`ERROR (code 301): Invalid character "/" after inline annotation
	in line 3 on file 
	> // inline comment
	--^`: {
				given: `{
	"foo": "bar", // foo comment
	// inline comment
	"fizz": "buzz"
}`,
			},

			`ERROR (code 802): Incorrect rule value type
	in line 2 on file 
	> {} // {type: @json}
	---------------^`: {
				given: `[
	{} // {type: @json}
]`,
			},

			`ERROR (code 301): Invalid character "@" key shortcut not allowed in annotation
	in line 2 on file 
	> {} // {@type: "foo"}
	---------^`: {
				given: `[
	{} // {@type: "foo"}
]`,
			},

			`ERROR (code 616): Date parsing error (parsing time "abc" as "2006-01-02": cannot parse "abc" as "2006")
	in line 2 on file 
	> "data": "abc" // {type: "date"}
	----------^`: {
				given: `{
	"data": "abc" // {type: "date"}
}`,
			},

			`ERROR (code 1302): Type "@petName" not found
	in line 3 on file 
	> @petName: @cat
	--^`: {
				types: map[string]string{
					"@cat": "{}",
				},
				given: `{
	"@notAShortCutKey": @cat,
	@petName: @cat
}`,
			},

			`ERROR (code 1301): Incorrect type of user type
	in line 1 on file 
	> 123 // {or: ["@cat", "@dog"]}
	--^`: {
				given: `123 // {or: ["@cat", "@dog"]}`,
				types: map[string]string{
					"@cat": `"cat"`,
					"@dog": `"dog"`,
				},
			},

			`ERROR (code 1117): The "minLength" constraint can't be used for the "email" type
	in line 1 on file 
	> "user@example.com" // {type: "email", minLength: 2}
	--^`: {
				given: `"user@example.com" // {type: "email", minLength: 2}`,
			},

			`ERROR (code 1117): The "maxLength" constraint can't be used for the "email" type
	in line 1 on file 
	> "user@example.com" // {type: "email", maxLength: 256}
	--^`: {
				given: `"user@example.com" // {type: "email", maxLength: 256}`,
			},

			`ERROR (code 1117): The "minLength" constraint can't be used for the "uri" type
	in line 1 on file 
	> "http://example.com" // {type: "uri", minLength: 2}
	--^`: {
				given: `"http://example.com" // {type: "uri", minLength: 2}`,
			},

			`ERROR (code 1117): The "maxLength" constraint can't be used for the "uri" type
	in line 1 on file 
	> "http://example.com" // {type: "uri", maxLength: 256}
	--^`: {
				given: `"http://example.com" // {type: "uri", maxLength: 256}`,
			},

			`ERROR (code 1117): The "minLength" constraint can't be used for the "date" type
	in line 1 on file 
	> "2022-02-27" // {type: "date", minLength: 2}
	--^`: {
				given: `"2022-02-27" // {type: "date", minLength: 2}`,
			},

			`ERROR (code 1117): The "maxLength" constraint can't be used for the "date" type
	in line 1 on file 
	> "2022-02-27" // {type: "date", maxLength: 256}
	--^`: {
				given: `"2022-02-27" // {type: "date", maxLength: 256}`,
			},

			`ERROR (code 1117): The "minLength" constraint can't be used for the "datetime" type
	in line 1 on file 
	> "2022-02-27T10:19:48+06:00" // {type: "datetime", minLength: 2}
	--^`: {
				given: `"2022-02-27T10:19:48+06:00" // {type: "datetime", minLength: 2}`,
			},

			`ERROR (code 1117): The "maxLength" constraint can't be used for the "datetime" type
	in line 1 on file 
	> "2022-02-27T10:19:48+06:00" // {type: "datetime", maxLength: 256}
	--^`: {
				given: `"2022-02-27T10:19:48+06:00" // {type: "datetime", maxLength: 256}`,
			},

			`ERROR (code 1117): The "minLength" constraint can't be used for the "uuid" type
	in line 1 on file 
	> "95f362d6-87df-4dd4-a948-9f84f65a3468" // {type: "uuid", minLength: 2}
	--^`: {
				given: `"95f362d6-87df-4dd4-a948-9f84f65a3468" // {type: "uuid", minLength: 2}`,
			},

			`ERROR (code 1117): The "maxLength" constraint can't be used for the "uuid" type
	in line 1 on file 
	> "95f362d6-87df-4dd4-a948-9f84f65a3468" // {type: "uuid", maxLength: 256}
	--^`: {
				given: `"95f362d6-87df-4dd4-a948-9f84f65a3468" // {type: "uuid", maxLength: 256}`,
			},

			`ERROR (code 1117): The "regex" constraint can't be used for the "uuid" type
	in line 1 on file 
	> "95f362d6-87df-4dd4-a948-9f84f65a3468" // {type: "uuid", regex: ".+"}
	--^`: {
				given: `"95f362d6-87df-4dd4-a948-9f84f65a3468" // {type: "uuid", regex: ".+"}`,
			},

			`ERROR (code 1117): The "const" constraint can't be used for the "any" type
	in line 1 on file 
	> 42 // {type: "any", const: true}
	--^`: {
				given: `42 // {type: "any", const: true}`,
			},

			`ERROR (code 1302): Type "@cat" not found
	in line 10 on file 
	> "bar": @cat
	---------^`: {
				given: `{
  "k1": 1,
  "k2": 2,
  "k3": 3,
  "k4": 4,
  "k5": 5,
  "k6": 6,
  "topFriends": {
    "foo": 42,
    "bar": @cat
  }
}`,
			},

			`ERROR (code 1302): Type "@petName" not found
	in line 10 on file 
	> @petName: @cat
	--^`: {
				given: `{
  "k1": 1,
  "k2": 2,
  "k3": 3,
  "k4": 4,
  "k5": 5,
  "k6": 6,
  "topFriends": {
    "foo": 42,
    @petName: @cat
  }
}`,
			},

			`ERROR (code 1117): The "minLength" constraint can't be used for the "float" type
	in line 1 on file 
	> 1.23 /* {precision: 2,
	--^`: {
				given: `1.23 /* {precision: 2,
                            minLength: 0,
                }*/`,
			},

			`ERROR (code 1117): The "minLength" constraint can't be used for the "decimal" type
	in line 1 on file 
	> 1.23 /* {type: "decimal", precision: 2,
	--^`: {
				given: `1.23 /* {type: "decimal", precision: 2,
                            minLength: 0,
                }*/`,
			},

			`ERROR (code 1117): The "precision" constraint can't be used for the "string" type
	in line 1 on file 
	> "user@example.com" // {precision: 2}
	--^`: {
				given: `"user@example.com" // {precision: 2}`,
			},

			`ERROR (code 1117): The "precision" constraint can't be used for the "email" type
	in line 1 on file 
	> "user@example.com" // {type: "email", precision: 2}
	--^`: {
				given: `"user@example.com" // {type: "email", precision: 2}`,
			},

			`ERROR (code 1117): The "precision" constraint can't be used for the "string" type
	in line 1 on file 
	> "2022-02-27" // {precision: 2}
	--^`: {
				given: `"2022-02-27" // {precision: 2}`,
			},

			`ERROR (code 1117): The "precision" constraint can't be used for the "date" type
	in line 1 on file 
	> "2022-02-27" // {type: "date", precision: 2}
	--^`: {
				given: `"2022-02-27" // {type: "date", precision: 2}`,
			},

			`ERROR (code 1117): The "precision" constraint can't be used for the "string" type
	in line 1 on file 
	> "2021-02-27T16:40:00+06:00" // {precision: 2}
	--^`: {
				given: `"2021-02-27T16:40:00+06:00" // {precision: 2}`,
			},

			`ERROR (code 1117): The "precision" constraint can't be used for the "datetime" type
	in line 1 on file 
	> "2021-02-27T16:40:00+06:00" // {type: "datetime", precision: 2}
	--^`: {
				given: `"2021-02-27T16:40:00+06:00" // {type: "datetime", precision: 2}`,
			},

			`ERROR (code 1117): The "precision" constraint can't be used for the "string" type
	in line 1 on file 
	> "https://example.com" // {precision: 2}
	--^`: {
				given: `"https://example.com" // {precision: 2}`,
			},

			`ERROR (code 1117): The "precision" constraint can't be used for the "uri" type
	in line 1 on file 
	> "https://example.com" // {type: "uri", precision: 2}
	--^`: {
				given: `"https://example.com" // {type: "uri", precision: 2}`,
			},

			`ERROR (code 1117): The "precision" constraint can't be used for the "string" type
	in line 1 on file 
	> "bea58dd8-5f05-4350-9705-18bcf10e70fa" // {precision: 2}
	--^`: {
				given: `"bea58dd8-5f05-4350-9705-18bcf10e70fa" // {precision: 2}`,
			},

			`ERROR (code 1117): The "precision" constraint can't be used for the "uuid" type
	in line 1 on file 
	> "bea58dd8-5f05-4350-9705-18bcf10e70fa" // {type: "uuid", precision: 2}
	--^`: {
				given: `"bea58dd8-5f05-4350-9705-18bcf10e70fa" // {type: "uuid", precision: 2}`,
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 2e2
	---^`: {
				given: `2e2`,
			},

			`ERROR (code 301): Invalid character "E" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 2E-2
	---^`: {
				given: `2E-2`,
			},

			`ERROR (code 301): Invalid character "E" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 2E+2
	---^`: {
				given: `2E+2`,
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 3.14e2
	------^`: {
				given: "3.14e2",
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 3.14e-2
	------^`: {
				given: "3.14e-2",
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 3.14e+2
	------^`: {
				given: "3.14e+2",
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 3.14e2 // {type: "decimal"}
	------^`: {
				given: `3.14e2 // {type: "decimal"}`,
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 3.14e-2 // {type: "decimal"}
	------^`: {
				given: `3.14e-2 // {type: "decimal"}`,
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 3.14e+2 // {type: "decimal"}
	------^`: {
				given: `3.14e+2 // {type: "decimal"}`,
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 2e2 // {type: "integer"}
	---^`: {
				given: `2e2 // {type: "integer"}`,
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 2e-2 // {type: "integer"}
	---^`: {
				given: `2e-2 // {type: "integer"}`,
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 2e+2 // {type: "integer"}
	---^`: {
				given: `2e+2 // {type: "integer"}`,
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 2e2 // {type: "float"}
	---^`: {
				given: `2e2 // {type: "float"}`,
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 2e-2 // {type: "float"}
	---^`: {
				given: `2e-2 // {type: "float"}`,
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 2e+2 // {type: "float"}
	---^`: {
				given: `2e+2 // {type: "float"}`,
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 2e2 // {type: "decimal"}
	---^`: {
				given: `2e2 // {type: "decimal"}`,
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 2e-2 // {type: "decimal"}
	---^`: {
				given: `2e-2 // {type: "decimal"}`,
			},

			`ERROR (code 301): Invalid character "e" isn't allowed 'cause not obvious it's a float or an integer
	in line 1 on file 
	> 2e+2 // {type: "decimal"}
	---^`: {
				given: `2e+2 // {type: "decimal"}`,
			},

			`ERROR (code 810): 42 value duplicates in "enum"
	in line 1 on file 
	> 42 // {enum: [42, 43, 42]}
	------------------------^`: {
				given: "42 // {enum: [42, 43, 42]}",
			},

			`ERROR (code 810): "bar" value duplicates in "enum"
	in line 1 on file 
	> "foo" // {enum: ["foo", "bar", "bar"]}
	---------------------------------^`: {
				given: `"foo" // {enum: ["foo", "bar", "bar"]}`,
			},

			`ERROR (code 302): Invalid character '2' in object key (inside comment)
	in line 2 on file 
	> "one": 1 // {min 25}
	-------------------^`: {
				given: `{
	"one": 1 // {min 25}
}`,
			},

			`ERROR (code 301): Invalid character "1" after object key
	in line 2 on file 
	> "one" 1
	--------^`: {
				given: `{
	"one" 1
}`,
			},

			`ERROR (code 1602): Enum rule "@enum" not found
	in line 1 on file 
	> "foo" // {enum: @enum}
	------------------^`: {
				given: `"foo" // {enum: @enum}`,
			},

			`ERROR (code 610): Does not match any of the enumeration values
	in line 1 on file 
	> 42 // {enum: @enum}
	--^`: {
				given: `42 // {enum: @enum}`,
				rules: map[string]string{
					"@enum": `["foo", "bar"]`,
				},
			},

			`ERROR (code 610): Does not match any of the enumeration values
	in line 2 on file 
	> "foo": 42 // {enum: @enum}
	---------^`: {
				given: `{
	"foo": 42 // {enum: @enum}
}`,
				rules: map[string]string{
					"@enum": `["foo", "bar"]`,
				},
			},

			`ERROR (code 806): An array or rule name was expected as a value for the "enum"
	in line 1 on file 
	> 42 // {enum: "@enum"}
	---------------^`: {
				given: `42 // {enum: "@enum"}`,
			},

			`ERROR (code 301): Invalid character "c" after object in inline annotation
	in line 2 on file 
	> "field": "value" // {optional: true} comment after rules without using dash
	---------------------------------------^`: {
				given: `{
    "field": "value" // {optional: true} comment after rules without using dash
  }`,
			},

			`ERROR (code 301): Invalid character "c" after object in multi-line annotation
	in line 4 on file 
	> comment after rules without using dash */
	--^`: {
				given: `{
    "field": "value" /*
                    {optional: true}
                    comment after rules without using dash */
  }`,
			},

			`ERROR (code 1115): Incompatible value of example and "type" rule (decimal)
	in line 1 on file 
	> "2" // {type: "decimal", precision: 2}
	--^`: {
				given: `"2" // {type: "decimal", precision: 2}`,
			},

			`ERROR (code 1115): Incompatible value of example and "type" rule (decimal)
	in line 1 on file 
	> 2 // {type: "decimal", precision: 2}
	--^`: {
				given: `2 // {type: "decimal", precision: 2}`,
			},

			`ERROR (code 1115): Incompatible value of example and "type" rule (email)
	in line 1 on file 
	> 10 // {type: "email"}
	--^`: {
				given: `10 // {type: "email"}`,
			},

			`ERROR (code 1115): Incompatible value of example and "type" rule (uri)
	in line 1 on file 
	> 10 // {type: "uri"}
	--^`: {
				given: `10 // {type: "uri"}`,
			},

			`ERROR (code 1115): Incompatible value of example and "type" rule (uuid)
	in line 1 on file 
	> 10 // {type: "uuid"}
	--^`: {
				given: `10 // {type: "uuid"}`,
			},

			`ERROR (code 1115): Incompatible value of example and "type" rule (date)
	in line 1 on file 
	> 10 // {type: "date"}
	--^`: {
				given: `10 // {type: "date"}`,
			},

			`ERROR (code 1115): Incompatible value of example and "type" rule (datetime)
	in line 1 on file 
	> 10 // {type: "datetime"}
	--^`: {
				given: `10 // {type: "datetime"}`,
			},

			`ERROR (code 402): Duplicate keys (@catId) in the schema
	in line 3 on file 
	> "@catId": 2
	--^`: {
				given: `{
  "@catId": 1,
  "@catId": 2
}`,
			},

			`ERROR (code 617): Value of constraint "min" should be less or equal to value of "max" constraint
	in line 1 on file 
	> 42 // {min: 45, max: 42}
	--^`: {
				given: "42 // {min: 45, max: 42}",
			},

			`ERROR (code 617): Value of constraint "minItems" should be less or equal to value of "maxItems" constraint
	in line 1 on file 
	> [ // {minItems: 2, maxItems: 1}
	--^`: {
				given: `[ // {minItems: 2, maxItems: 1}
    1,2
  ]`,
			},

			`ERROR (code 617): Value of constraint "minLength" should be less or equal to value of "maxLength" constraint
	in line 1 on file 
	> "foo" // {minLength: 2, maxLength: 1}
	--^`: {
				given: `"foo" // {minLength: 2, maxLength: 1}`,
			},

			`ERROR (code 602): Invalid value for "min" = 43 constraint 
	in line 1 on file 
	> 42 // {min: 43, max: 44}
	--^`: {
				given: "42 // {min: 43, max: 44}",
			},

			`ERROR (code 602): Invalid value for "max" = 41 constraint 
	in line 1 on file 
	> 42 // {min: 30, max: 41}
	--^`: {
				given: "42 // {min: 30, max: 41}",
			},

			`ERROR (code 608): The number of array elements does not match the "minItems" rule
	in line 1 on file 
	> [ // {minItems: 2, maxItems: 3}
	--^`: {
				given: `[ // {minItems: 2, maxItems: 3}
    1
  ]`,
			},

			`ERROR (code 609): The number of array elements does not match the "maxItems" rule
	in line 1 on file 
	> [ // {minItems: 1, maxItems: 2}
	--^`: {
				given: `[ // {minItems: 1, maxItems: 2}
    1,2,3
  ]`,
			},

			`ERROR (code 603): Invalid string length for "minLength" = "4" constraint
	in line 1 on file 
	> "foo" // {minLength: 4, maxLength: 5}
	--^`: {
				given: `"foo" // {minLength: 4, maxLength: 5}`,
			},

			`ERROR (code 603): Invalid string length for "maxLength" = "2" constraint
	in line 1 on file 
	> "foo" // {minLength: 1, maxLength: 2}
	--^`: {
				given: `"foo" // {minLength: 1, maxLength: 2}`,
			},

			`ERROR (code 1304): Key shortcut "@foo" should be string but "integer" given
	in line 2 on file 
	> @foo: 42
	--^`: {
				given: `{
	@foo: 42
}`,
				types: map[string]string{
					"@foo": "42",
				},
			},

			`ERROR (code 1304): Key shortcut "@foo" should be string but "float" given
	in line 2 on file 
	> @foo: 42
	--^`: {
				given: `{
	@foo: 42
}`,
				types: map[string]string{
					"@foo": "3.14",
				},
			},

			`ERROR (code 1304): Key shortcut "@foo" should be string but "boolean" given
	in line 2 on file 
	> @foo: 42
	--^`: {
				given: `{
	@foo: 42
}`,
				types: map[string]string{
					"@foo": "true",
				},
			},

			`ERROR (code 1304): Key shortcut "@foo" should be string but "null" given
	in line 2 on file 
	> @foo: 42
	--^`: {
				given: `{
	@foo: 42
}`,
				types: map[string]string{
					"@foo": "null",
				},
			},

			`ERROR (code 1304): Key shortcut "@foo" should be string but "array" given
	in line 2 on file 
	> @foo: 42
	--^`: {
				given: `{
	@foo: 42
}`,
				types: map[string]string{
					"@foo": "[1,2,3]",
				},
			},

			`ERROR (code 1304): Key shortcut "@foo" should be string but "object" given
	in line 2 on file 
	> @foo: 42
	--^`: {
				given: `{
	@foo: 42
}`,
				types: map[string]string{
					"@foo": `{"fizz": "buzz"}`,
				},
			},

			`ERROR (code 402): Duplicate keys (@catId) in the schema
	in line 4 on file 
	> "@catId": 3,
	--^`: {
				given: `{
  "@catId": 1,
  @catId: 2,
  "@catId": 3,
  @catId: 4
}`,
				types: map[string]string{
					"@catId": `"12" // A cat's id.`,
				},
			},

			`ERROR (code 810): "a" value duplicates in "enum"
	in line 1 on file 
	> "a" // {enum: ["a", "a"]}
	----------------------^`: {
				given: "\"a\" // {enum: [\"a\", \"\u0061\"]}",
			},

			`ERROR (code 810): "b" value duplicates in "enum"
	in line 1 on file 
	> "b" // {enum: ["b", "b"]}
	----------------------^`: {
				given: `"b" // {enum: ["b", "\u0062"]}`,
			},
		}

		for expected, c := range cc {
			t.Run(expected, func(t *testing.T) {
				s := MustNew("", c.given)

				for n, b := range c.rules {
					require.NoError(t, s.AddRule(n, enum.MustNew(n, b)))
				}

				for n, b := range c.types {
					err := s.AddType(n, MustNew(n, b))
					if err != nil {
						require.EqualError(t, err, expected)
					}
				}

				err := s.Check()
				assert.EqualError(t, err, expected)
			})
		}

		t.Run("req.jschema.rules.type.reference 0.2", func(t *testing.T) {
			cc := map[string]string{
				`ERROR (code 1107): You cannot specify child node if you use a type reference
	in line 2 on file 
	> "myCat": { // {type: "@cat"}
	-----------^`: `{
	"myCat": { // {type: "@cat"}
		"id"  : 123,
		"name": "Tom"
	}
}`,
				`ERROR (code 1107): You cannot specify child node if you use a type reference
	in line 2 on file 
	> "myCatList": [ // {type: "@catList"}
	---------------^`: `{
					"myCatList": [ // {type: "@catList"}
						@cat
					]
				}`,
				`ERROR (code 1107): You cannot specify child node if you use a type reference
	in line 1 on file 
	> {} // {type: "@foo"}
	--^`: `{} // {type: "@foo"}`,
				`ERROR (code 1107): You cannot specify child node if you use a type reference
	in line 1 on file 
	> [] // {type: "@foo"}
	--^`: `[] // {type: "@foo"}`,
			}

			for expected, schema := range cc {
				t.Run(expected, func(t *testing.T) {
					assert.EqualError(t, MustNew("", schema).Check(), expected)
				})
			}
		})

		t.Run("req.jschema.rules.or 0.2", func(t *testing.T) {
			cc := map[string]string{
				`ERROR (code 1108): You cannot specify child node if you use a "or" rule
	in line 2 on file 
	> "myPet1": { // {or: ["@cat", "@dog"]}
	------------^`: `{
	"myPet1": { // {or: ["@cat", "@dog"]}
		"id": 1,
		"name": "Tom"
	}
}`,

				`ERROR (code 1108): You cannot specify child node if you use a "or" rule
	in line 2 on file 
	> "myPets": [ // {or: ["@catList", "@dogList"]}
	------------^`: `{
	"myPets": [ // {or: ["@catList", "@dogList"]}
		@cat
	]
}`,

				`ERROR (code 501): Duplicate "types" rule
	in line 2 on file 
	> "myPet4" : @cat | @dog // {or: ["@cat", "@dog"]}
	---------------------------------^`: `{
	"myPet4" : @cat | @dog // {or: ["@cat", "@dog"]}
}`,

				`ERROR (code 1108): You cannot specify child node if you use a "or" rule
	in line 2 on file 
	> "id": {} // {or: ["@cat", "@dog"]}
	--------^`: `{
	"id": {} // {or: ["@cat", "@dog"]}
}`,

				`ERROR (code 1108): You cannot specify child node if you use a "or" rule
	in line 2 on file 
	> "myPet3" : @cat // {or: ["@cat", "@dog"]}  # --ERROR! It is wrong.
	-------------^`: `{
	"myPet3" : @cat // {or: ["@cat", "@dog"]}  # --ERROR! It is wrong.
}`,
			}

			for expected, schema := range cc {
				t.Run(expected, func(t *testing.T) {
					assert.EqualError(t, MustNew("", schema).Check(), expected)
				})
			}
		})

		t.Run("invalid recursion", func(t *testing.T) {
			cc := map[string]struct {
				given string
				types map[string]string
			}{
				"Infinity recursion detected @main -> @bar -> @main": {
					given: `{
	"foo": @bar
}`,
					types: map[string]string{
						"@bar": `{
	"bar": @main
}`,
					},
				},

				"Infinity recursion detected @main -> @fizz -> @main": {
					given: `{
	"foo": @fizz | @buzz
}`,
					types: map[string]string{
						"@fizz": `{
	"fizz": @main
}`,
						"@buzz": `{
	"buzz": @main
}`,
					},
				},

				"Infinity recursion detected @main -> @main": {
					given: `{ // {allOf: ["@foo", "@bar"]}
}`,
					types: map[string]string{
						"@foo": `{
	"foo": 42
}`,
						"@bar": `{
	"bar": @main
}`,
					},
				},
			}

			for expected, c := range cc {
				t.Run(expected, func(t *testing.T) {
					s := MustNew("@main", c.given)

					for n, b := range c.types {
						require.NoError(t, s.AddType(n, MustNew(n, b)))
					}
					require.NoError(t, s.AddType("@main", s))

					err := s.Check()
					assert.EqualError(t, err, expected)
				})
			}
		})
	})
}

func TestSchema_Validate(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]struct {
			schema string
			types  map[string]string
			jsons  []string
		}{
			"object": {
				schema: `
{
	"foo": 1,
	"bar": "string"
}`,
				jsons: []string{`
{
	"foo": 42,
	"bar": "fizz"
}`},
			},

			"allOf": {
				schema: `
{ // {allOf: "@aaa"}
	"bbb": 222
}
`,
				types: map[string]string{
					"@aaa": `{"aaa": 111}`,
				},
				jsons: []string{
					`{"aaa": 1, "bbb": 2}`,
					`{"aaa": 1}`,
					`{"bbb": 2}`,
					`{}`,
				},
			},

			"user type nullable": {
				schema: `{
	"foo": 1 // {type: "@bar", nullable: true}
}`,
				types: map[string]string{
					"@bar": "123",
				},
				jsons: []string{
					`{"foo": 42}`,
					`{"foo": null}`,
				},
			},

			"shortcut nullable": {
				schema: `{
	"foo": @bar // {nullable: true}
}`,
				types: map[string]string{
					"@bar": "123",
				},
				jsons: []string{
					`{"foo": 24}`,
					`{"foo": null}`,
				},
			},

			"or nullable": {
				schema: `{
	"foo": @fizz | @buzz // {nullable: true}
}`,
				types: map[string]string{
					"@fizz": "[]",
					"@buzz": "{}",
				},
				jsons: []string{
					`{"foo": []}`,
					`{"foo": {}}`,
					`{"foo": null}`,
				},
			},

			"enum nullable": {
				schema: `{
	"foo": 1 // {enum: [1, 2, 3], nullable: true}
}`,
				jsons: []string{
					`{"foo": 1}`,
					`{"foo": 2}`,
					`{"foo": 3}`,
					`{"foo": null}`,
				},
			},

			"or with types (objects)": {
				schema: `42 /* {or: [
	{type: "boolean"},
	{type: "integer"},
	{type: "float"},
	{type: "null"},
	{type: "string"},
	{type: "@foo"}
]} */`,
				types: map[string]string{
					"@foo": `"foo-1" // {regex: "foo-[0-9]+"}`,
				},
				jsons: []string{
					"42",
					"3.14",
					"true",
					"null",
					"false",
					`"foo-42"`,
					`"fizz"`,
				},
			},

			"or with types (mixed)": {
				schema: `42 /* {or: [
	{type: "boolean"},
	"integer",
	{type: "float"},
	"null",
	{type: "string"},
	"@foo"
]} */`,
				types: map[string]string{
					"@foo": `"foo-1" // {regex: "foo-[0-9]+"}`,
				},
				jsons: []string{
					"42",
					"3.14",
					"true",
					"null",
					"false",
					`"foo-42"`,
					`"fizz"`,
				},
			},

			"or with types (flat)": {
				schema: `42 /* {or: [
	"boolean",
	"integer",
	"float",
	"null",
	"string",
	"@foo"
]} */`,
				types: map[string]string{
					"@foo": `"foo-1" // {regex: "foo-[0-9]+"}`,
				},
				jsons: []string{
					"42",
					"3.14",
					"true",
					"null",
					"false",
					`"foo-42"`,
					`"fizz"`,
				},
			},

			"Or without type": {
				schema: `{
	"foo": 123 /* {or: [
		{min: 100},
		{type: "string"}
	]} */
}`,
				jsons: []string{
					`{"foo": 1000}`,
					`{"foo": "bar"}`,
				},
			},
		}

		for name, c := range cc {
			t.Run(name, func(t *testing.T) {
				schema := MustNew("schema", c.schema, KeysAreOptionalByDefault())

				for n, s := range c.types {
					require.NoError(t, schema.AddType(n, MustNew(s, s, KeysAreOptionalByDefault())))
				}

				for _, s := range c.jsons {
					t.Run(s, func(t *testing.T) {
						err := schema.Validate(json.MustNew("json", s))
						require.NoError(t, err)
					})
				}
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]struct {
			schema string
			types  map[string]string
			json   string
		}{
			`ERROR (code 1302): Type "@int" not found
	in line 2 on file schema
	> "aaa": 111 // {type: "@int"}
	---------^`: {
				schema: `{
		"aaa": 111 // {type: "@int"}
	}`,
			},

			`ERROR (code 1301): Incorrect type of user type
	in line 2 on file schema
	> "aaa": 111 // {type: "@int"}
	---------^`: {
				schema: `{
		"aaa": 111 // {type: "@int"}
	}`,
				types: map[string]string{
					"@int": `"abc"`,
				},
			},

			`ERROR (code 204): None of the rules in the "OR" set has been validated
	in line 1 on file json
	> {"foo": 10}
	----------^`: {
				schema: `{
	"foo": 123 /* {or: [
		{min: 100},
		{type: "string"}
	]} */
}`,
				json: `{"foo": 10}`,
			},

			`ERROR (code 204): None of the rules in the "OR" set has been validated
	in line 1 on file json
	> {"foo": true}
	----------^`: {
				schema: `{
	"foo": 123 /* {or: [
		{min: 100},
		{type: "string"}
	]} */
}`,
				json: `{"foo": true}`,
			},

			`ERROR (code 1117): The "precision" constraint can't be used for the "float" type
	in line 1 on file schema
	> 1.1 // {type: "float", precision: 2}
	--^`: {
				schema: `1.1 // {type: "float", precision: 2}`,
				json:   "3.14",
			},
		}

		for expected, c := range cc {
			t.Run(expected, func(t *testing.T) {
				schema := MustNew("schema", c.schema, KeysAreOptionalByDefault())

				for n, s := range c.types {
					require.NoError(t, schema.AddType(n, MustNew(s, s, KeysAreOptionalByDefault())))
				}

				err := schema.Validate(json.MustNew("json", c.json))
				assert.EqualError(t, err, expected)
			})
		}

		t.Run("not a JSON document", func(t *testing.T) {
			err := MustNew("schema", "42").Validate(&mocks.Document{})
			assert.EqualError(t, err, "support only JSON documents, but got *mocks.Document")
		})
	})
}

func TestSchema_GetAST(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]struct {
			expected jschema.ASTNode
			types    map[string]string
			rules    map[string]string
		}{
			"@foo": {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeShortcut,
					SchemaType: "@foo",
					Value:      "@foo",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"type": {
								TokenType:  jschema.TokenTypeShortcut,
								Value:      "@foo",
								Properties: &jschema.RuleASTNodes{},
								Source:     jschema.RuleASTNodeSourceGenerated,
							},
						},
						[]string{"type"},
					),
				},
				types: map[string]string{
					"@foo": `"foo"`,
				},
			},

			"   @foo   ": {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeShortcut,
					SchemaType: "@foo",
					Value:      "@foo",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"type": {
								TokenType:  jschema.TokenTypeShortcut,
								Value:      "@foo",
								Properties: &jschema.RuleASTNodes{},
								Source:     jschema.RuleASTNodeSourceGenerated,
							},
						},
						[]string{"type"},
					),
				},
				types: map[string]string{
					"@foo": `"foo"`,
				},
			},

			"   @foo | @bar   ": {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeShortcut,
					SchemaType: string(jschema.SchemaTypeMixed),
					Value:      "@foo | @bar",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"or": {
								TokenType:  jschema.TokenTypeArray,
								Properties: &jschema.RuleASTNodes{},
								Items: []jschema.RuleASTNode{
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "@foo",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceGenerated,
									},
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "@bar",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceGenerated,
									},
								},
								Source: jschema.RuleASTNodeSourceGenerated,
							},
						},
						[]string{"or"},
					),
				},
				types: map[string]string{
					"@foo": `"foo"`,
					"@bar": `"bar"`,
				},
			},

			`{
				"data": "abc" /* {
					or: [
						"@foo",
						{type: "@bar"}
					]
				} */
			}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Rules:      &jschema.RuleASTNodes{},
					Children: []jschema.ASTNode{
						{
							Key:        "data",
							TokenType:  jschema.TokenTypeString,
							SchemaType: string(jschema.SchemaTypeMixed),
							Value:      "abc",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"or": {
										TokenType:  jschema.TokenTypeArray,
										Properties: &jschema.RuleASTNodes{},
										Items: []jschema.RuleASTNode{
											{
												TokenType:  jschema.TokenTypeString,
												Value:      "@foo",
												Properties: &jschema.RuleASTNodes{},
												Source:     jschema.RuleASTNodeSourceManual,
											},
											{
												TokenType: jschema.TokenTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															TokenType:  jschema.TokenTypeShortcut,
															Value:      "@bar",
															Properties: &jschema.RuleASTNodes{},
															Source:     jschema.RuleASTNodeSourceManual,
														},
													},
													[]string{"type"},
												),
												Source: jschema.RuleASTNodeSourceManual,
											},
										},
										Source: jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"or"},
							),
						},
					},
				},
				types: map[string]string{
					"@foo": `"foo"`,
					"@bar": `"bar"`,
				},
			},

			`{
				"data": "abc" /* {
					or: [
						{type: "@foo"},
						{type: "@bar"}
					]
				} */
			}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Rules:      &jschema.RuleASTNodes{},
					Children: []jschema.ASTNode{
						{
							Key:        "data",
							TokenType:  jschema.TokenTypeString,
							SchemaType: string(jschema.SchemaTypeMixed),
							Value:      "abc",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"or": {
										TokenType:  jschema.TokenTypeArray,
										Properties: &jschema.RuleASTNodes{},
										Items: []jschema.RuleASTNode{
											{
												TokenType: jschema.TokenTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															TokenType:  jschema.TokenTypeShortcut,
															Value:      "@foo",
															Properties: &jschema.RuleASTNodes{},
															Source:     jschema.RuleASTNodeSourceManual,
														},
													},
													[]string{"type"},
												),
												Source: jschema.RuleASTNodeSourceManual,
											},
											{
												TokenType: jschema.TokenTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															TokenType:  jschema.TokenTypeShortcut,
															Value:      "@bar",
															Properties: &jschema.RuleASTNodes{},
															Source:     jschema.RuleASTNodeSourceManual,
														},
													},
													[]string{"type"},
												),
												Source: jschema.RuleASTNodeSourceManual,
											},
										},
										Source: jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"or"},
							),
						},
					},
				},
				types: map[string]string{
					"@foo": `"foo"`,
					"@bar": `"bar"`,
				},
			},

			`{
				"data": "abc" /* {
					or: [
						{type: "string" , maxLength: 3},
						{type: "integer", min: 0}
					]
				} */
			}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "data",
							TokenType:  jschema.TokenTypeString,
							SchemaType: string(jschema.SchemaTypeMixed),
							Value:      "abc",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"or": {
										TokenType:  jschema.TokenTypeArray,
										Properties: &jschema.RuleASTNodes{},
										Items: []jschema.RuleASTNode{
											{
												TokenType: jschema.TokenTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															TokenType:  jschema.TokenTypeString,
															Value:      "string",
															Properties: &jschema.RuleASTNodes{},
															Source:     jschema.RuleASTNodeSourceManual,
														},
														"maxLength": {
															TokenType:  jschema.TokenTypeNumber,
															Value:      "3",
															Properties: &jschema.RuleASTNodes{},
															Source:     jschema.RuleASTNodeSourceManual,
														},
													},
													[]string{"type", "maxLength"},
												),
												Source: jschema.RuleASTNodeSourceManual,
											},
											{
												TokenType: jschema.TokenTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															TokenType:  jschema.TokenTypeString,
															Value:      "integer",
															Properties: &jschema.RuleASTNodes{},
															Source:     jschema.RuleASTNodeSourceManual,
														},
														"min": {
															TokenType:  jschema.TokenTypeNumber,
															Value:      "0",
															Properties: &jschema.RuleASTNodes{},
															Source:     jschema.RuleASTNodeSourceManual,
														},
													},
													[]string{"type", "min"},
												),
												Source: jschema.RuleASTNodeSourceManual,
											},
										},
										Source: jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"or"},
							),
						},
					},
					Rules: &jschema.RuleASTNodes{},
				},
			},

			`1 // {type: "mixed", or: ["@foo", "@bar"]}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeNumber,
					SchemaType: string(jschema.SchemaTypeMixed),
					Value:      "1",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"type": {
								TokenType:  jschema.TokenTypeString,
								Value:      "mixed",
								Properties: &jschema.RuleASTNodes{},
								Source:     jschema.RuleASTNodeSourceManual,
							},
							"or": {
								TokenType:  jschema.TokenTypeArray,
								Properties: &jschema.RuleASTNodes{},
								Items: []jschema.RuleASTNode{
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "@foo",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "@bar",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								Source: jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"type", "or"},
					),
				},
				types: map[string]string{
					"@foo": `42`,
					"@bar": `"bar"`,
				},
			},

			`1 // {type: "mixed", or: ["@fizz", "@buzz"]}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeNumber,
					SchemaType: string(jschema.SchemaTypeMixed),
					Value:      "1",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"type": {
								TokenType:  jschema.TokenTypeString,
								Value:      "mixed",
								Properties: &jschema.RuleASTNodes{},
								Source:     jschema.RuleASTNodeSourceManual,
							},
							"or": {
								TokenType:  jschema.TokenTypeArray,
								Properties: &jschema.RuleASTNodes{},
								Items: []jschema.RuleASTNode{
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "@fizz",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "@buzz",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								Source: jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"type", "or"},
					),
				},
				types: map[string]string{
					"@fizz": `42`,
					"@buzz": `"buzz"`,
				},
			},

			`"section0" // {regex: "section[0-9]"}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeString,
					SchemaType: string(jschema.SchemaTypeString),
					Value:      "section0",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"regex": {
								TokenType:  jschema.TokenTypeString,
								Value:      "section[0-9]",
								Properties: &jschema.RuleASTNodes{},
								Source:     jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"regex"},
					),
				},
			},

			`
123 /*
        {min: 0}
      */
`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeNumber,
					SchemaType: string(jschema.SchemaTypeInteger),
					Value:      "123",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"min": {
								TokenType:  jschema.TokenTypeNumber,
								Value:      "0",
								Properties: &jschema.RuleASTNodes{},
								Source:     jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"min"},
					),
				},
			},

			`{
  "id1": 1, // {type: "@id", nullable: true}
  "id2": @id, // {nullable: true}
  "id3": @id1 | @id2, // {nullable: true}
  "size": 1, // {enum: [1,2,3], nullable: true}
  "choice": 1 // {or: [{type: "integer"}, {type: "string"}]}
}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id1",
							TokenType:  jschema.TokenTypeNumber,
							SchemaType: "@id",
							Value:      "1",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"type": {
										TokenType:  jschema.TokenTypeShortcut,
										Properties: &jschema.RuleASTNodes{},
										Value:      "@id",
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Properties: &jschema.RuleASTNodes{},
										Value:      "true",
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"type", "nullable"},
							),
						},
						{
							Key:        "id2",
							TokenType:  jschema.TokenTypeShortcut,
							SchemaType: "@id",
							Value:      "@id",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"type": {
										TokenType:  jschema.TokenTypeShortcut,
										Value:      "@id",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceGenerated,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Properties: &jschema.RuleASTNodes{},
										Value:      "true",
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"type", "nullable"},
							),
						},
						{
							Key:        "id3",
							TokenType:  jschema.TokenTypeShortcut,
							SchemaType: string(jschema.SchemaTypeMixed),
							Value:      "@id1 | @id2",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"or": {
										TokenType:  jschema.TokenTypeArray,
										Properties: &jschema.RuleASTNodes{},
										Items: []jschema.RuleASTNode{
											{
												TokenType:  jschema.TokenTypeString,
												Value:      "@id1",
												Properties: &jschema.RuleASTNodes{},
												Source:     jschema.RuleASTNodeSourceGenerated,
											},
											{
												TokenType:  jschema.TokenTypeString,
												Value:      "@id2",
												Properties: &jschema.RuleASTNodes{},
												Source:     jschema.RuleASTNodeSourceGenerated,
											},
										},
										Source: jschema.RuleASTNodeSourceGenerated,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Properties: &jschema.RuleASTNodes{},
										Value:      "true",
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"or", "nullable"},
							),
						},
						{
							Key:        "size",
							TokenType:  jschema.TokenTypeNumber,
							SchemaType: string(jschema.SchemaTypeEnum),
							Value:      "1",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"enum": {
										TokenType:  jschema.TokenTypeArray,
										Properties: &jschema.RuleASTNodes{},
										Items: []jschema.RuleASTNode{
											{
												TokenType:  jschema.TokenTypeNumber,
												Value:      "1",
												Properties: &jschema.RuleASTNodes{},
												Source:     jschema.RuleASTNodeSourceManual,
											},
											{
												TokenType:  jschema.TokenTypeNumber,
												Value:      "2",
												Properties: &jschema.RuleASTNodes{},
												Source:     jschema.RuleASTNodeSourceManual,
											},
											{
												TokenType:  jschema.TokenTypeNumber,
												Value:      "3",
												Properties: &jschema.RuleASTNodes{},
												Source:     jschema.RuleASTNodeSourceManual,
											},
										},
										Source: jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Properties: &jschema.RuleASTNodes{},
										Value:      "true",
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"enum", "nullable"},
							),
						},
						{
							Key:        "choice",
							TokenType:  jschema.TokenTypeNumber,
							SchemaType: string(jschema.SchemaTypeMixed),
							Value:      "1",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"or": {
										TokenType:  jschema.TokenTypeArray,
										Properties: &jschema.RuleASTNodes{},
										Items: []jschema.RuleASTNode{
											{
												TokenType: jschema.TokenTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															TokenType:  jschema.TokenTypeString,
															Value:      "integer",
															Properties: &jschema.RuleASTNodes{},
															Source:     jschema.RuleASTNodeSourceManual,
														},
													},
													[]string{"type"},
												),
												Source: jschema.RuleASTNodeSourceManual,
											},
											{
												TokenType: jschema.TokenTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															TokenType:  jschema.TokenTypeString,
															Value:      "string",
															Properties: &jschema.RuleASTNodes{},
															Source:     jschema.RuleASTNodeSourceManual,
														},
													},
													[]string{"type"},
												),
												Source: jschema.RuleASTNodeSourceManual,
											},
										},
										Source: jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"or"},
							),
						},
					},
					Rules: &jschema.RuleASTNodes{},
				},
				types: map[string]string{
					"@id":  "1",
					"@id1": "2",
					"@id2": "3",
				},
			},

			"[]  // {minItems: 0} - Description": {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeArray,
					SchemaType: string(jschema.SchemaTypeArray),
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"minItems": {
								TokenType:  jschema.TokenTypeNumber,
								Value:      "0",
								Properties: &jschema.RuleASTNodes{},
								Source:     jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"minItems"},
					),
					Comment: "Description",
				},
			},

			`{
	"foo": [1],
	"bar": 42 // number
}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "foo",
							TokenType:  jschema.TokenTypeArray,
							SchemaType: string(jschema.SchemaTypeArray),
							Rules:      &jschema.RuleASTNodes{},
							Children: []jschema.ASTNode{
								{
									TokenType:  jschema.TokenTypeNumber,
									SchemaType: string(jschema.SchemaTypeInteger),
									Value:      "1",
									Rules:      &jschema.RuleASTNodes{},
								},
							},
						},
						{
							Key:        "bar",
							TokenType:  jschema.TokenTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "42",
							Rules:      &jschema.RuleASTNodes{},
							Comment:    "number",
						},
					},
					Rules: &jschema.RuleASTNodes{},
				},
			},

			`[ // Comment
	1
]`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeArray,
					SchemaType: string(jschema.SchemaTypeArray),
					Rules:      &jschema.RuleASTNodes{},
					Children: []jschema.ASTNode{
						{
							TokenType:  jschema.TokenTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "1",
							Rules:      &jschema.RuleASTNodes{},
						},
					},
					Comment: "Comment",
				},
			},

			"[] // Comment": {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeArray,
					SchemaType: string(jschema.SchemaTypeArray),
					Rules:      &jschema.RuleASTNodes{},
					Comment:    "Comment",
				},
			},

			`[
	[],
	2 // Annotation
]`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeArray,
					SchemaType: string(jschema.SchemaTypeArray),
					Rules:      &jschema.RuleASTNodes{},
					Children: []jschema.ASTNode{
						{
							TokenType:  jschema.TokenTypeArray,
							SchemaType: string(jschema.SchemaTypeArray),
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							TokenType:  jschema.TokenTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "2",
							Rules:      &jschema.RuleASTNodes{},
							Comment:    "Annotation",
						},
					},
				},
			},

			`"A" // {or: ["string", "integer"]}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeString,
					SchemaType: string(jschema.SchemaTypeMixed),

					Value: "A",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"or": {
								TokenType:  jschema.TokenTypeArray,
								Properties: &jschema.RuleASTNodes{},
								Items: []jschema.RuleASTNode{
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "string",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "integer",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								Source: jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"or"},
					),
				},
			},

			`{
	"foo": 123 /* {or: [
		{min: 100},
		{type: "string"}
	]} */
}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "foo",
							TokenType:  jschema.TokenTypeNumber,
							SchemaType: string(jschema.SchemaTypeMixed),
							Value:      "123",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"or": {
										TokenType:  jschema.TokenTypeArray,
										Properties: &jschema.RuleASTNodes{},
										Items: []jschema.RuleASTNode{
											{
												TokenType: jschema.TokenTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"min": {
															TokenType:  jschema.TokenTypeNumber,
															Value:      "100",
															Properties: &jschema.RuleASTNodes{},
															Source:     jschema.RuleASTNodeSourceManual,
														},
													},
													[]string{"min"},
												),
												Source: jschema.RuleASTNodeSourceManual,
											},
											{
												TokenType: jschema.TokenTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															TokenType:  jschema.TokenTypeString,
															Value:      "string",
															Properties: &jschema.RuleASTNodes{},
															Source:     jschema.RuleASTNodeSourceManual,
														},
													},
													[]string{"type"},
												),
												Source: jschema.RuleASTNodeSourceManual,
											},
										},
										Source: jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"or"},
							),
						},
					},
					Rules: &jschema.RuleASTNodes{},
				},
			},

			`{
  "enabled": { // {additionalProperties: true, nullable: false}
  },
  "disabled": { // {additionalProperties: false, nullable: false}
  },
  "string": { // {additionalProperties: "string", nullable: false}
  },
  "integer": { // {additionalProperties: "integer", nullable: false}
  },
  "float": { // {additionalProperties: "float", nullable: false}
  },
  "decimal": { // {additionalProperties: "decimal", nullable: false}
  },
  "boolean": { // {additionalProperties: "boolean", nullable: false}
  },
  "object": { // {additionalProperties: "object", nullable: false}
  },
  "array": { // {additionalProperties: "array", nullable: false}
  },
  "null": { // {additionalProperties: "null", nullable: false}
  },
  "email": { // {additionalProperties: "email", nullable: false}
  },
  "uri": { // {additionalProperties: "uri", nullable: false}
  },
  "uuid": { // {additionalProperties: "uuid", nullable: false}
  },
  "date": { // {additionalProperties: "date", nullable: false}
  },
  "datetime": { // {additionalProperties: "datetime", nullable: false}
  },
  "enum": { // {additionalProperties: "enum", nullable: false}
  },
  "mixed": { // {additionalProperties: "mixed", nullable: false}
  },
  "any": { // {additionalProperties: "any", nullable: false}
  },
  "userType": { // {additionalProperties: "@cat", nullable: false}
  }
}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "enabled",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "true",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "disabled",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "string",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "string",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "integer",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "integer",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "float",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "float",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "decimal",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "decimal",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "boolean",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "boolean",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "object",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "object",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "array",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "array",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "null",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "null",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "email",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "email",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "uri",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "uri",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "uuid",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "uuid",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "date",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "date",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "datetime",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "datetime",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "enum",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "enum",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "mixed",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "mixed",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "any",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "any",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
						{
							Key:        "userType",
							TokenType:  jschema.TokenTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										TokenType:  jschema.TokenTypeString,
										Value:      "@cat",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										TokenType:  jschema.TokenTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"additionalProperties", "nullable"},
							),
						},
					},
					Rules: &jschema.RuleASTNodes{},
				},
				types: map[string]string{
					"@cat": `"cat"`,
				},
			},

			`{
	@fooKey: @foo
}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:           "@fooKey",
							IsKeyShortcut: true,
							TokenType:     jschema.TokenTypeShortcut,
							SchemaType:    "@foo",
							Value:         "@foo",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"type": {
										TokenType:  jschema.TokenTypeShortcut,
										Value:      "@foo",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceGenerated,
									},
								},
								[]string{"type"},
							),
						},
					},
					Rules: &jschema.RuleASTNodes{},
				},
				types: map[string]string{
					"@fooKey": `"key"`,
					"@foo":    `"foo"`,
				},
			},

			`"foo" /* {or: [
                  {type: "string"},
                  {type: "boolean"},
                  {type: "integer"},
                  {type: "float"},
                  {type: "object"},
                  {type: "array"},
                  {type: "decimal", precision: 1}
                ]} 
            */`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeString,
					SchemaType: string(jschema.SchemaTypeMixed),
					Value:      "foo",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"or": {
								TokenType:  jschema.TokenTypeArray,
								Properties: &jschema.RuleASTNodes{},
								Items: []jschema.RuleASTNode{
									{
										TokenType: jschema.TokenTypeObject,
										Properties: jschema.NewRuleASTNodes(
											map[string]jschema.RuleASTNode{
												"type": {
													TokenType:  jschema.TokenTypeString,
													Value:      "string",
													Properties: &jschema.RuleASTNodes{},
													Source:     jschema.RuleASTNodeSourceManual,
												},
											},
											[]string{"type"},
										),
										Source: jschema.RuleASTNodeSourceManual,
									},
									{
										TokenType: jschema.TokenTypeObject,
										Properties: jschema.NewRuleASTNodes(
											map[string]jschema.RuleASTNode{
												"type": {
													TokenType:  jschema.TokenTypeString,
													Value:      "boolean",
													Properties: &jschema.RuleASTNodes{},
													Source:     jschema.RuleASTNodeSourceManual,
												},
											},
											[]string{"type"},
										),
										Source: jschema.RuleASTNodeSourceManual,
									},
									{
										TokenType: jschema.TokenTypeObject,
										Properties: jschema.NewRuleASTNodes(
											map[string]jschema.RuleASTNode{
												"type": {
													TokenType:  jschema.TokenTypeString,
													Value:      "integer",
													Properties: &jschema.RuleASTNodes{},
													Source:     jschema.RuleASTNodeSourceManual,
												},
											},
											[]string{"type"},
										),
										Source: jschema.RuleASTNodeSourceManual,
									},
									{
										TokenType: jschema.TokenTypeObject,
										Properties: jschema.NewRuleASTNodes(
											map[string]jschema.RuleASTNode{
												"type": {
													TokenType:  jschema.TokenTypeString,
													Value:      "float",
													Properties: &jschema.RuleASTNodes{},
													Source:     jschema.RuleASTNodeSourceManual,
												},
											},
											[]string{"type"},
										),
										Source: jschema.RuleASTNodeSourceManual,
									},
									{
										TokenType: jschema.TokenTypeObject,
										Properties: jschema.NewRuleASTNodes(
											map[string]jschema.RuleASTNode{
												"type": {
													TokenType:  jschema.TokenTypeString,
													Value:      "object",
													Properties: &jschema.RuleASTNodes{},
													Source:     jschema.RuleASTNodeSourceManual,
												},
											},
											[]string{"type"},
										),
										Source: jschema.RuleASTNodeSourceManual,
									},
									{
										TokenType: jschema.TokenTypeObject,
										Properties: jschema.NewRuleASTNodes(
											map[string]jschema.RuleASTNode{
												"type": {
													TokenType:  jschema.TokenTypeString,
													Value:      "array",
													Properties: &jschema.RuleASTNodes{},
													Source:     jschema.RuleASTNodeSourceManual,
												},
											},
											[]string{"type"},
										),
										Source: jschema.RuleASTNodeSourceManual,
									},
									{
										TokenType: jschema.TokenTypeObject,
										Properties: jschema.NewRuleASTNodes(
											map[string]jschema.RuleASTNode{
												"type": {
													TokenType:  jschema.TokenTypeString,
													Value:      "decimal",
													Properties: &jschema.RuleASTNodes{},
													Source:     jschema.RuleASTNodeSourceManual,
												},
												"precision": {
													TokenType:  jschema.TokenTypeNumber,
													Value:      "1",
													Properties: &jschema.RuleASTNodes{},
													Source:     jschema.RuleASTNodeSourceManual,
												},
											},
											[]string{"type", "precision"},
										),
										Source: jschema.RuleASTNodeSourceManual,
									},
								},
								Source: jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"or"},
					),
				},
			},

			`1.2 // {precision: 2}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeNumber,
					SchemaType: string(jschema.SchemaTypeDecimal),

					Value: "1.2",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"precision": {
								TokenType:  jschema.TokenTypeNumber,
								Value:      "2",
								Properties: &jschema.RuleASTNodes{},
								Source:     jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"precision"},
					),
				},
			},

			`"a" // {or: ["string", "integer"]}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeString,
					SchemaType: string(jschema.SchemaTypeMixed),

					Value: "a",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"or": {
								TokenType:  jschema.TokenTypeArray,
								Properties: &jschema.RuleASTNodes{},
								Items: []jschema.RuleASTNode{
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "string",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "integer",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								Source: jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"or"},
					),
				},
			},

			`"cat" /*
            {enum: [
              "cat", // The cat
              "dog", // The dog
              "pig", // The pig
              "frog" // The frog
            ]}
        */`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeString,
					SchemaType: string(jschema.SchemaTypeEnum),

					Value: "cat",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"enum": {
								TokenType:  jschema.TokenTypeArray,
								Properties: &jschema.RuleASTNodes{},
								Items: []jschema.RuleASTNode{
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "cat",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
										Comment:    "The cat",
									},
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "dog",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
										Comment:    "The dog",
									},
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "pig",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
										Comment:    "The pig",
									},
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "frog",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
										Comment:    "The frog",
									},
								},
								Source: jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"enum"},
					),
				},
			},

			`"foo" // {type: "string"} - annotation # should not be a comment in AST node`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeString,
					SchemaType: string(jschema.SchemaTypeString),

					Value: "foo",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"type": {
								TokenType:  jschema.TokenTypeString,
								Properties: &jschema.RuleASTNodes{},
								Value:      "string",
								Source:     jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"type"},
					),
					Comment: "annotation",
				},
			},

			`"#" // {regex: "#"} - annotation # comment`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeString,
					SchemaType: string(jschema.SchemaTypeString),

					Value: "#",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"regex": {
								TokenType:  jschema.TokenTypeString,
								Properties: &jschema.RuleASTNodes{},
								Value:      "#",
								Source:     jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"regex"},
					),
					Comment: "annotation",
				},
			},

			`"#" // {enum: ["#", "##"]} - annotation # comment`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeString,
					SchemaType: string(jschema.SchemaTypeEnum),

					Value: "#",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"enum": {
								TokenType:  jschema.TokenTypeArray,
								Properties: &jschema.RuleASTNodes{},
								Items: []jschema.RuleASTNode{
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "#",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									{
										TokenType:  jschema.TokenTypeString,
										Value:      "##",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								Source: jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"enum"},
					),
					Comment: "annotation",
				},
			},

			`{
  "id": 5,
  "name": "John" # single-line COMMENT
}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							TokenType:  jschema.TokenTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							TokenType:  jschema.TokenTypeString,
							SchemaType: string(jschema.SchemaTypeString),
							Value:      "John",
							Rules:      &jschema.RuleASTNodes{},
						},
					},
					Rules: &jschema.RuleASTNodes{},
				},
			},

			`{
  "id": 5,
  "name": "John"
  ###
  block
  COMMENT
  ###
}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							TokenType:  jschema.TokenTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							TokenType:  jschema.TokenTypeString,
							SchemaType: string(jschema.SchemaTypeString),
							Value:      "John",
							Rules:      &jschema.RuleASTNodes{},
						},
					},
					Rules: &jschema.RuleASTNodes{},
				},
			},

			`{
  "id": 5,
  "name": "John" /*
  # comment
*/
}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							TokenType:  jschema.TokenTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							TokenType:  jschema.TokenTypeString,
							SchemaType: string(jschema.SchemaTypeString),
							Value:      "John",
							Rules:      &jschema.RuleASTNodes{},
							Comment:    "# comment",
						},
					},
					Rules: &jschema.RuleASTNodes{},
				},
			},

			`{
  "id": 5,
  "name": "John" /* {type: "string"} - annotation
  # comment
*/
}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							TokenType:  jschema.TokenTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							TokenType:  jschema.TokenTypeString,
							SchemaType: string(jschema.SchemaTypeString),
							Value:      "John",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"type": {
										TokenType:  jschema.TokenTypeString,
										Value:      "string",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"type"},
							),
							Comment: `annotation
  # comment`,
						},
					},
					Rules: &jschema.RuleASTNodes{},
				},
			},

			`{
  "id": 5,
  "name": "John" /* {type: "string"} - annotation # comment
*/
}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							TokenType:  jschema.TokenTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							TokenType:  jschema.TokenTypeString,
							SchemaType: string(jschema.SchemaTypeString),
							Value:      "John",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"type": {
										TokenType:  jschema.TokenTypeString,
										Value:      "string",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"type"},
							),
							Comment: `annotation # comment`,
						},
					},
					Rules: &jschema.RuleASTNodes{},
				},
			},

			`{
  "id": 5,
  "name": "John" // {type: "string"} - annotation # comment
}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							TokenType:  jschema.TokenTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							TokenType:  jschema.TokenTypeString,
							SchemaType: string(jschema.SchemaTypeString),
							Value:      "John",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"type": {
										TokenType:  jschema.TokenTypeString,
										Value:      "string",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"type"},
							),
							Comment: `annotation`,
						},
					},
					Rules: &jschema.RuleASTNodes{},
				},
			},

			`{
  "id": 5,
  "name": "John" /*
  ###
  block
  COMMENT
  ###
*/
}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							TokenType:  jschema.TokenTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							TokenType:  jschema.TokenTypeString,
							SchemaType: string(jschema.SchemaTypeString),
							Value:      "John",
							Rules:      &jschema.RuleASTNodes{},
							Comment: `###
  block
  COMMENT
  ###`,
						},
					},
					Rules: &jschema.RuleASTNodes{},
				},
			},

			`{
  "id": 5,
  "name": "John" /* {type: "string"} - annotation
  ###
  block
  COMMENT
  ###
*/
}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							TokenType:  jschema.TokenTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							TokenType:  jschema.TokenTypeString,
							SchemaType: string(jschema.SchemaTypeString),
							Value:      "John",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"type": {
										TokenType:  jschema.TokenTypeString,
										Value:      "string",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
								},
								[]string{"type"},
							),
							Comment: `annotation
  ###
  block
  COMMENT
  ###`,
						},
					},
					Rules: &jschema.RuleASTNodes{},
				},
			},

			`# {
#  "id": 5,
#  "name": "John"
# }`: {
				expected: jschema.ASTNode{
					Rules: &jschema.RuleASTNodes{},
				},
			},

			`"foo" // {enum: @enum}`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeString,
					SchemaType: string(jschema.SchemaTypeEnum),
					Value:      "foo",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"enum": {
								TokenType:  jschema.TokenTypeShortcut,
								Value:      "@enum",
								Properties: &jschema.RuleASTNodes{},
								Source:     jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"enum"},
					),
				},
				rules: map[string]string{
					"@enum": `[
// Comment 1
"foo", // Comment 2
// Comment 3
"bar"  // Comment 4
// Comment 5
]`,
				},
			},

			`"foo" /* {
	type: "string"
} - comment
*/`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeString,
					SchemaType: string(jschema.SchemaTypeString),
					Value:      "foo",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"type": {
								TokenType:  jschema.TokenTypeString,
								Value:      "string",
								Properties: &jschema.RuleASTNodes{},
								Source:     jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"type"},
					),
					Comment: "comment",
				},
			},

			`"foo" /* {
	type: "string"
} - multi
line
	comment
*/`: {
				expected: jschema.ASTNode{
					TokenType:  jschema.TokenTypeString,
					SchemaType: string(jschema.SchemaTypeString),
					Value:      "foo",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"type": {
								TokenType:  jschema.TokenTypeString,
								Value:      "string",
								Properties: &jschema.RuleASTNodes{},
								Source:     jschema.RuleASTNodeSourceManual,
							},
						},
						[]string{"type"},
					),
					Comment: "multi\nline\n\tcomment",
				},
			},
		}

		for given, c := range cc {
			t.Run(given, func(t *testing.T) {
				s := MustNew("", given)

				for n, r := range c.rules {
					require.NoError(t, s.AddRule(n, enum.MustNew(n, r)))
				}

				for n, c := range c.types {
					require.NoError(t, s.AddType(n, MustNew(n, c)))
				}

				actual, err := s.GetAST()
				require.NoError(t, err)
				assert.Equalf(
					t,
					c.expected,
					actual,
					fmt.Sprintf("Expected: %s\nActual: %s", spew.Sdump(c.expected), spew.Sdump(actual)),
				)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]struct {
			schema string
			types  map[string]string
		}{
			`ERROR (code 102): Unknown type "foo"
	in line 1 on file 
	> 42 // {type: "foo"}
	--^`: {
				schema: `42 // {type: "foo"}`,
			},

			`ERROR (code 616): Date parsing error (parsing time "abc" as "2006-01-02": cannot parse "abc" as "2006")
	in line 2 on file 
	> "data": "abc" // {type: "date"}
	----------^`: {
				schema: `{
	"data": "abc" // {type: "date"}
}`,
			},

			`ERROR (code 301): Invalid character "," non-space byte after top-level value
	in line 1 on file 
	> @pig, // {or: ["@dog", "@pig"]}
	------^`: {
				schema: `@pig, // {or: ["@dog", "@pig"]}`,
			},

			`ERROR (code 304): Annotation not allowed here
	in line 2 on file 
	> "ids": [1] // Ids
	-------------^`: {
				schema: `{
	"ids": [1] // Ids
}`,
			},

			`ERROR (code 304): Annotation not allowed here
	in line 3 on file 
	> 1] // Ids
	-----^`: {
				schema: `{
	"ids": [
1] // Ids
}`,
			},

			`ERROR (code 304): Annotation not allowed here
	in line 4 on file 
	> ] // Ids
	----^`: {
				schema: `{
	"ids": [
	1
] // Ids
}`,
			},

			`ERROR (code 1108): You cannot specify child node if you use a "or" rule
	in line 2 on file 
	> "foo" : @fizz // {or: ["@fizz", "@buzz"]}
	----------^`: {
				schema: `{
	"foo" : @fizz // {or: ["@fizz", "@buzz"]}
}`,
			},

			`ERROR (code 1108): You cannot specify child node if you use a "or" rule
	in line 2 on file 
	> "foo": {} // {or: ["@fizz", "@buzz"]}
	---------^`: {
				schema: `{
	"foo": {} // {or: ["@fizz", "@buzz"]}
}`,
			},

			`ERROR (code 1107): You cannot specify child node if you use a type reference
	in line 2 on file 
	> "foo" : @fizz // {type: "@fizz"}
	----------^`: {
				schema: `{
	"foo" : @fizz // {type: "@fizz"}
}`,
			},

			`ERROR (code 1107): You cannot specify child node if you use a type reference
	in line 2 on file 
	> "foo": {} // {type: "@fizz"}
	---------^`: {
				schema: `{
	"foo": {} // {type: "@fizz"}
}`,
			},

			`ERROR (code 303): Unexpected end of file
	in line 1 on file 
	> 1.
	---^`: {
				schema: "1.",
			},

			`ERROR (code 301): Invalid character "\n" after decimal point in numeric literal
	in line 1 on file 
	> 1.
	----^`: {
				schema: "1.\n",
			},

			`ERROR (code 1117): The "precision" constraint can't be used for the "float" type
	in line 1 on file 
	> 1.1 // {type: "float", precision: 2}
	--^`: {
				schema: `1.1 // {type: "float", precision: 2}`,
			},

			`ERROR (code 1302): Type "@foo" not found
	in line 1 on file 
	> @foo
	--^`: {
				schema: "@foo",
			},

			`ERROR (code 1301): Incorrect type of user type
	in line 1 on file 
	> 123 // {or: ["@cat", "@dog"]}
	--^`: {
				schema: `123 // {or: ["@cat", "@dog"]}`,
				types: map[string]string{
					"@cat": `"cat"`,
					"@dog": `"dog"`,
				},
			},
		}

		for expected, c := range cc {
			t.Run(expected, func(t *testing.T) {
				s := MustNew("", c.schema)

				for n, c := range c.types {
					require.NoError(t, s.AddType(n, MustNew(n, c)))
				}

				_, err := s.GetAST()
				assert.EqualError(t, err, expected)
			})
		}
	})
}

func TestSchema_UsedUserTypes(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string][]string{
			"@foo":        {"@foo"},
			"@foo | @bar": {"@foo", "@bar"},
			`{
	"foo": @foo,
	"bar": {
		"fizz": @bar | @fizz,
		"buzz": 42, // {type: "@buzz"}
		"foobar": 42 // {or: ["@foobar", {type: "@fizzbuzz"}]}
	},
	"fizzbuzz": [
		@foobarfizzbuzz
	],
	"scalar": 3.14, // {type: "decimal", precision: 2}
	"scalar_or": 42, // {or: ["string", {type: "integer"}]}
	"allof": { // {allOf: "@base"}
	},
	"allof_array": { // {allOf: ["@base1", "@base2"]}
	},
	"@notAShortcut": 42,
	@shortcut: 42
}`: {
				"@foo",
				"@bar",
				"@fizz",
				"@buzz",
				"@foobar",
				"@fizzbuzz",
				"@foobarfizzbuzz",
				"@base",
				"@base1",
				"@base2",
				"@shortcut",
			},
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				ss, err := MustNew("", given).UsedUserTypes()
				require.NoError(t, err)
				assert.ElementsMatch(t, expected, ss)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		_, err := MustNew("", "foo").UsedUserTypes()
		assert.EqualError(t, err, `ERROR (code 301): Invalid character "o" in literal false (expecting 'a')
	in line 1 on file 
	> foo
	---^`)
	})
}

func TestSchema_Build(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		err := MustNew("", "42").Build()
		assert.NoError(t, err)
	})

	t.Run("negative", func(t *testing.T) {
		err := MustNew("", "foo").Build()
		assert.EqualError(t, err, `ERROR (code 301): Invalid character "o" in literal false (expecting 'a')
	in line 1 on file 
	> foo
	---^`)
	})
}

func TestSchema_buildASTNode(t *testing.T) {
	t.Run("root node nil", func(t *testing.T) {
		s := &Schema{
			inner: &internalSchema.Schema{},
		}

		n := s.buildASTNode()
		assert.Equal(t, jschema.ASTNode{
			Rules: &jschema.RuleASTNodes{},
		}, n)
	})

	t.Run("root node isn't nil", func(t *testing.T) {
		newSchema := func(rootNode internalSchema.Node) *Schema {
			inner := internalSchema.New()
			inner.SetRootNode(rootNode)
			return &Schema{
				inner: &inner,
			}
		}

		t.Run("positive", func(t *testing.T) {
			expected := jschema.ASTNode{
				TokenType: jschema.TokenTypeString,
				Rules:     &jschema.RuleASTNodes{},
			}

			root := &schemaMocks.Node{}
			root.On("ASTNode").Return(expected, nil)

			s := newSchema(root)

			n := s.buildASTNode()
			assert.Equal(t, expected, n)
		})

		t.Run("negative", func(t *testing.T) {
			root := &schemaMocks.Node{}
			root.On("ASTNode").Return(jschema.ASTNode{}, stdErrors.New("fake error"))

			s := newSchema(root)

			assert.PanicsWithError(t, "fake error", func() {
				s.buildASTNode()
			})
		})
	})
}
