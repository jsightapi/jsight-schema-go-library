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
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/internal/mocks"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/loader"
	schemaMocks "github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/mocks"
	internalSchema "github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema"
	"github.com/jsightapi/jsight-schema-go-library/notations/regex"
	"github.com/jsightapi/jsight-schema-go-library/rules/enum"
)

func ExampleSchema() {
	s := New("root", `{"foo": @Fizz,"bar": @Buzz}`)

	l, err := s.Len()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	fmt.Println(l)

	err = s.AddType("@Fizz", New("fizz", `{"fizz": 1}`))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	err = s.AddType("@Buzz", New("buzz", `{"buzz": 2}`))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	err = s.Check()
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}

	err = s.Validate(json.New("json", `{"foo":{"fizz":42},"bar":{"buzz":42}}`))
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
				l, err := New("foo", given).Len()
				require.NoError(t, err)
				assert.Equal(t, int(expected), int(l))
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		_, err := New("foo", "invalid").Len()
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
		s := New("foo", `[
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
		t.Skip() // todo fix in SERV-8
		s := New("schema", `
{
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
	"recursion": @recursion
}
`)

		require.NoError(t, s.AddRule("@enum", enum.New("@enum", `[
	1,
	2,
	3
]`)))
		require.NoError(t, s.AddType("@foo", New("@foo", `{
	"foo": 42
}`)))
		require.NoError(t, s.AddType("@bar", New("@bar", `{
	"bar": 42
}`)))
		require.NoError(t, s.AddType("@recursion", New("@recursion", `{
	"recursion": @recursion
}`)))

		actual, err := s.Example()
		require.NoError(t, err)
		assert.JSONEq(t, `{
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
	"enum_rule": 2
}`,
			string(actual),
		)
	})

	t.Run("negative", func(t *testing.T) {
		_, err := New("schema", "invalid").
			Example()
		assert.EqualError(t, err, `ERROR (code 301): Invalid character "i" looking for beginning of value
	in line 1 on file schema
	> invalid
	--^`)
	})
}

func Benchmark_buildExample(b *testing.B) {
	node := loader.NewSchemaForSdk(
		fs.NewFile("", `{
	"foo": "bar",
	"fizz": [
		1,
		2,
		3
	],
	"buzz": {
		"foo": [
			{"bar": 1},
			{"bar": 2}
		],
		"bar": {
			"fizz": 42,
			"buzz": [1, 2, 3]
		},
		"fizz": 1, // {or: ["string", "integer"]}
		"buzz": 2
	}
}`),
		false,
	).
		RootNode()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = buildExample(node)
	}
}

func TestSchema_AddType(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		t.Run("jschema", func(t *testing.T) {
			root := New("", `{"foo": @foo}`)
			typ := New("", "123")
			err := root.AddType("@foo", typ)
			require.NoError(t, err)

			require.NotNil(t, root.inner)
			assert.Equal(t, typ.inner, root.inner.Type("@foo"))
		})

		t.Run("regex", func(t *testing.T) {
			root := New("", `{"foo": @foo}`)
			typ := regex.New("", "/foo-\\d/")
			err := root.AddType("@foo", typ)
			require.NoError(t, err)

			require.NotNil(t, root.inner)
		})
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("invalid schema", func(t *testing.T) {
			err := New("", "42").AddType("invalid", nil)
			assert.EqualError(t, err, "schema should be JSight or Regex schema, but <nil> given")
		})

		t.Run("invalid schema name", func(t *testing.T) {
			err := New("", "42").AddType("invalid", New("invalid", "42"))
			assert.EqualError(t, err, "Invalid schema name (invalid)")
		})
	})
}

func TestSchema_AddRule(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		const name = "foo"
		r := mocks.NewRule(t)
		r.On("Check").Return(nil)
		s := New("", "content")

		err := s.AddRule(name, r)

		require.NoError(t, err)
		assert.Len(t, s.rules, 1)
		assert.Contains(t, s.rules, name)
		assert.Same(t, r, s.rules[name])
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("already compiled", func(t *testing.T) {
			s := New("foo", "content")
			s.inner = &internalSchema.Schema{}

			err := s.AddRule("foo", mocks.NewRule(t))

			assert.EqualError(t, err, "schema is already compiled")
			assert.Len(t, s.rules, 0)
		})

		t.Run("nil rule", func(t *testing.T) {
			s := New("", "content")

			err := s.AddRule("", nil)

			assert.EqualError(t, err, "rule is nil")
			assert.Len(t, s.rules, 0)
		})

		t.Run("invalid rule", func(t *testing.T) {
			r := mocks.NewRule(t)
			r.On("Check").Return(stdErrors.New("fake error"))
			s := New("", "content")

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
		}

		for content, c := range cc {
			t.Run(content, func(t *testing.T) {
				s := New("@main", content)

				for n, c := range c.enums {
					require.NoError(t, s.AddRule(n, enum.New(n, c)))
				}

				for n, c := range c.types {
					require.NoError(t, s.AddType(n, New(n, c)))
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
		}

		for expected, c := range cc {
			t.Run(expected, func(t *testing.T) {
				s := New("", c.given)

				for n, b := range c.rules {
					require.NoError(t, s.AddRule(n, enum.New(n, b)))
				}

				for n, b := range c.types {
					require.NoError(t, s.AddType(n, New(n, b)))
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
					assert.EqualError(t, New("", schema).Check(), expected)
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
					assert.EqualError(t, New("", schema).Check(), expected)
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
					s := New("@main", c.given)

					for n, b := range c.types {
						require.NoError(t, s.AddType(n, New(n, b)))
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
				schema := New("schema", c.schema, KeysAreOptionalByDefault())

				for n, s := range c.types {
					require.NoError(t, schema.AddType(n, New(s, s, KeysAreOptionalByDefault())))
				}

				for _, s := range c.jsons {
					t.Run(s, func(t *testing.T) {
						err := schema.Validate(json.New("json", s))
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
				schema := New("schema", c.schema, KeysAreOptionalByDefault())

				for n, s := range c.types {
					require.NoError(t, schema.AddType(n, New(s, s, KeysAreOptionalByDefault())))
				}

				err := schema.Validate(json.New("json", c.json))
				assert.EqualError(t, err, expected)
			})
		}

		t.Run("not a JSON document", func(t *testing.T) {
			err := New("schema", "42").Validate(&mocks.Document{})
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
					JSONType:   jschema.JSONTypeShortcut,
					SchemaType: "@foo",
					Value:      "@foo",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"type": {
								JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeShortcut,
					SchemaType: "@foo",
					Value:      "@foo",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"type": {
								JSONType:   "string",
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
					JSONType:   jschema.JSONTypeShortcut,
					SchemaType: string(jschema.SchemaTypeMixed),
					Value:      "@foo | @bar",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"or": {
								JSONType:   jschema.JSONTypeArray,
								Properties: &jschema.RuleASTNodes{},
								Items: []jschema.RuleASTNode{
									{
										JSONType:   jschema.JSONTypeString,
										Value:      "@foo",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceGenerated,
									},
									{
										JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Rules:      &jschema.RuleASTNodes{},
					Children: []jschema.ASTNode{
						{
							Key:        "data",
							JSONType:   jschema.JSONTypeString,
							SchemaType: string(jschema.SchemaTypeMixed),
							Value:      "abc",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"or": {
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
												JSONType: jschema.JSONTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Rules:      &jschema.RuleASTNodes{},
					Children: []jschema.ASTNode{
						{
							Key:        "data",
							JSONType:   jschema.JSONTypeString,
							SchemaType: string(jschema.SchemaTypeMixed),
							Value:      "abc",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"or": {
										JSONType:   jschema.JSONTypeArray,
										Properties: &jschema.RuleASTNodes{},
										Items: []jschema.RuleASTNode{
											{
												JSONType: jschema.JSONTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															JSONType:   jschema.JSONTypeString,
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
												JSONType: jschema.JSONTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "data",
							JSONType:   jschema.JSONTypeString,
							SchemaType: string(jschema.SchemaTypeMixed),
							Value:      "abc",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"or": {
										JSONType:   jschema.JSONTypeArray,
										Properties: &jschema.RuleASTNodes{},
										Items: []jschema.RuleASTNode{
											{
												JSONType: jschema.JSONTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															JSONType:   jschema.JSONTypeString,
															Value:      "string",
															Properties: &jschema.RuleASTNodes{},
															Source:     jschema.RuleASTNodeSourceManual,
														},
														"maxLength": {
															JSONType:   jschema.JSONTypeNumber,
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
												JSONType: jschema.JSONTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															JSONType:   jschema.JSONTypeString,
															Value:      "integer",
															Properties: &jschema.RuleASTNodes{},
															Source:     jschema.RuleASTNodeSourceManual,
														},
														"min": {
															JSONType:   jschema.JSONTypeNumber,
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
					JSONType:   jschema.JSONTypeNumber,
					SchemaType: string(jschema.SchemaTypeMixed),
					Value:      "1",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"type": {
								JSONType:   jschema.JSONTypeString,
								Value:      "mixed",
								Properties: &jschema.RuleASTNodes{},
								Source:     jschema.RuleASTNodeSourceManual,
							},
							"or": {
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
					JSONType:   jschema.JSONTypeNumber,
					SchemaType: string(jschema.SchemaTypeMixed),
					Value:      "1",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"type": {
								JSONType:   jschema.JSONTypeString,
								Value:      "mixed",
								Properties: &jschema.RuleASTNodes{},
								Source:     jschema.RuleASTNodeSourceManual,
							},
							"or": {
								JSONType:   jschema.JSONTypeArray,
								Properties: &jschema.RuleASTNodes{},
								Items: []jschema.RuleASTNode{
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
					JSONType:   jschema.JSONTypeString,
					SchemaType: string(jschema.SchemaTypeString),
					Value:      "section0",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"regex": {
								JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeNumber,
					SchemaType: string(jschema.SchemaTypeInteger),
					Value:      "123",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"min": {
								JSONType:   jschema.JSONTypeNumber,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id1",
							JSONType:   jschema.JSONTypeNumber,
							SchemaType: "@id",
							Value:      "1",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"type": {
										JSONType:   jschema.JSONTypeString,
										Properties: &jschema.RuleASTNodes{},
										Value:      "@id",
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeShortcut,
							SchemaType: "@id",
							Value:      "@id",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"type": {
										JSONType:   jschema.JSONTypeString,
										Value:      "@id",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceGenerated,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeShortcut,
							SchemaType: string(jschema.SchemaTypeMixed),
							Value:      "@id1 | @id2",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"or": {
										JSONType:   jschema.JSONTypeArray,
										Properties: &jschema.RuleASTNodes{},
										Items: []jschema.RuleASTNode{
											{
												JSONType:   jschema.JSONTypeString,
												Value:      "@id1",
												Properties: &jschema.RuleASTNodes{},
												Source:     jschema.RuleASTNodeSourceGenerated,
											},
											{
												JSONType:   jschema.JSONTypeString,
												Value:      "@id2",
												Properties: &jschema.RuleASTNodes{},
												Source:     jschema.RuleASTNodeSourceGenerated,
											},
										},
										Source: jschema.RuleASTNodeSourceGenerated,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeNumber,
							SchemaType: string(jschema.SchemaTypeEnum),
							Value:      "1",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"enum": {
										JSONType:   jschema.JSONTypeArray,
										Properties: &jschema.RuleASTNodes{},
										Items: []jschema.RuleASTNode{
											{
												JSONType:   jschema.JSONTypeNumber,
												Value:      "1",
												Properties: &jschema.RuleASTNodes{},
												Source:     jschema.RuleASTNodeSourceManual,
											},
											{
												JSONType:   jschema.JSONTypeNumber,
												Value:      "2",
												Properties: &jschema.RuleASTNodes{},
												Source:     jschema.RuleASTNodeSourceManual,
											},
											{
												JSONType:   jschema.JSONTypeNumber,
												Value:      "3",
												Properties: &jschema.RuleASTNodes{},
												Source:     jschema.RuleASTNodeSourceManual,
											},
										},
										Source: jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeNumber,
							SchemaType: string(jschema.SchemaTypeMixed),
							Value:      "1",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"or": {
										JSONType:   jschema.JSONTypeArray,
										Properties: &jschema.RuleASTNodes{},
										Items: []jschema.RuleASTNode{
											{
												JSONType: jschema.JSONTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															JSONType:   jschema.JSONTypeString,
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
												JSONType: jschema.JSONTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeArray,
					SchemaType: string(jschema.SchemaTypeArray),
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"minItems": {
								JSONType:   jschema.JSONTypeNumber,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "foo",
							JSONType:   jschema.JSONTypeArray,
							SchemaType: string(jschema.SchemaTypeArray),
							Rules:      &jschema.RuleASTNodes{},
							Children: []jschema.ASTNode{
								{
									JSONType:   jschema.JSONTypeNumber,
									SchemaType: string(jschema.SchemaTypeInteger),
									Value:      "1",
									Rules:      &jschema.RuleASTNodes{},
								},
							},
						},
						{
							Key:        "bar",
							JSONType:   jschema.JSONTypeNumber,
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
					JSONType:   jschema.JSONTypeArray,
					SchemaType: string(jschema.SchemaTypeArray),
					Rules:      &jschema.RuleASTNodes{},
					Children: []jschema.ASTNode{
						{
							JSONType:   jschema.JSONTypeNumber,
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
					JSONType:   jschema.JSONTypeArray,
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
					JSONType:   jschema.JSONTypeArray,
					SchemaType: string(jschema.SchemaTypeArray),
					Rules:      &jschema.RuleASTNodes{},
					Children: []jschema.ASTNode{
						{
							JSONType:   jschema.JSONTypeArray,
							SchemaType: string(jschema.SchemaTypeArray),
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							JSONType:   jschema.JSONTypeNumber,
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
					JSONType:   jschema.JSONTypeString,
					SchemaType: string(jschema.SchemaTypeMixed),

					Value: "A",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"or": {
								JSONType:   jschema.JSONTypeArray,
								Properties: &jschema.RuleASTNodes{},
								Items: []jschema.RuleASTNode{
									{
										JSONType:   jschema.JSONTypeString,
										Value:      "string",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									{
										JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "foo",
							JSONType:   jschema.JSONTypeNumber,
							SchemaType: string(jschema.SchemaTypeMixed),
							Value:      "123",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"or": {
										JSONType:   jschema.JSONTypeArray,
										Properties: &jschema.RuleASTNodes{},
										Items: []jschema.RuleASTNode{
											{
												JSONType: jschema.JSONTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"min": {
															JSONType:   jschema.JSONTypeNumber,
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
												JSONType: jschema.JSONTypeObject,
												Properties: jschema.NewRuleASTNodes(
													map[string]jschema.RuleASTNode{
														"type": {
															JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "enabled",
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeBoolean,
										Value:      "true",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeBoolean,
										Value:      "false",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "string",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "integer",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "float",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "decimal",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "boolean",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "object",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "array",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "null",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "email",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "uri",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "uuid",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "date",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "datetime",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "enum",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "mixed",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "any",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
							JSONType:   jschema.JSONTypeObject,
							SchemaType: string(jschema.SchemaTypeObject),
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"additionalProperties": {
										JSONType:   jschema.JSONTypeString,
										Value:      "@cat",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									"nullable": {
										JSONType:   jschema.JSONTypeBoolean,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:           "@fooKey",
							IsKeyShortcut: true,
							JSONType:      jschema.JSONTypeShortcut,
							SchemaType:    "@foo",
							Value:         "@foo",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"type": {
										JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeString,
					SchemaType: string(jschema.SchemaTypeMixed),
					Value:      "foo",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"or": {
								JSONType:   jschema.JSONTypeArray,
								Properties: &jschema.RuleASTNodes{},
								Items: []jschema.RuleASTNode{
									{
										JSONType: jschema.JSONTypeObject,
										Properties: jschema.NewRuleASTNodes(
											map[string]jschema.RuleASTNode{
												"type": {
													JSONType:   jschema.JSONTypeString,
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
										JSONType: jschema.JSONTypeObject,
										Properties: jschema.NewRuleASTNodes(
											map[string]jschema.RuleASTNode{
												"type": {
													JSONType:   jschema.JSONTypeString,
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
										JSONType: jschema.JSONTypeObject,
										Properties: jschema.NewRuleASTNodes(
											map[string]jschema.RuleASTNode{
												"type": {
													JSONType:   jschema.JSONTypeString,
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
										JSONType: jschema.JSONTypeObject,
										Properties: jschema.NewRuleASTNodes(
											map[string]jschema.RuleASTNode{
												"type": {
													JSONType:   jschema.JSONTypeString,
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
										JSONType: jschema.JSONTypeObject,
										Properties: jschema.NewRuleASTNodes(
											map[string]jschema.RuleASTNode{
												"type": {
													JSONType:   jschema.JSONTypeString,
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
										JSONType: jschema.JSONTypeObject,
										Properties: jschema.NewRuleASTNodes(
											map[string]jschema.RuleASTNode{
												"type": {
													JSONType:   jschema.JSONTypeString,
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
										JSONType: jschema.JSONTypeObject,
										Properties: jschema.NewRuleASTNodes(
											map[string]jschema.RuleASTNode{
												"type": {
													JSONType:   jschema.JSONTypeString,
													Value:      "decimal",
													Properties: &jschema.RuleASTNodes{},
													Source:     jschema.RuleASTNodeSourceManual,
												},
												"precision": {
													JSONType:   jschema.JSONTypeNumber,
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
					JSONType:   jschema.JSONTypeNumber,
					SchemaType: string(jschema.SchemaTypeDecimal),

					Value: "1.2",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"precision": {
								JSONType:   jschema.JSONTypeNumber,
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
					JSONType:   jschema.JSONTypeString,
					SchemaType: string(jschema.SchemaTypeMixed),

					Value: "a",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"or": {
								JSONType:   jschema.JSONTypeArray,
								Properties: &jschema.RuleASTNodes{},
								Items: []jschema.RuleASTNode{
									{
										JSONType:   jschema.JSONTypeString,
										Value:      "string",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									{
										JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeString,
					SchemaType: string(jschema.SchemaTypeEnum),

					Value: "cat",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"enum": {
								JSONType:   jschema.JSONTypeArray,
								Properties: &jschema.RuleASTNodes{},
								Items: []jschema.RuleASTNode{
									{
										JSONType:   jschema.JSONTypeString,
										Value:      "cat",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
										Comment:    "The cat",
									},
									{
										JSONType:   jschema.JSONTypeString,
										Value:      "dog",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
										Comment:    "The dog",
									},
									{
										JSONType:   jschema.JSONTypeString,
										Value:      "pig",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
										Comment:    "The pig",
									},
									{
										JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeString,
					SchemaType: string(jschema.SchemaTypeString),

					Value: "foo",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"type": {
								JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeString,
					SchemaType: string(jschema.SchemaTypeString),

					Value: "#",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"regex": {
								JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeString,
					SchemaType: string(jschema.SchemaTypeEnum),

					Value: "#",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"enum": {
								JSONType:   jschema.JSONTypeArray,
								Properties: &jschema.RuleASTNodes{},
								Items: []jschema.RuleASTNode{
									{
										JSONType:   jschema.JSONTypeString,
										Value:      "#",
										Properties: &jschema.RuleASTNodes{},
										Source:     jschema.RuleASTNodeSourceManual,
									},
									{
										JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							JSONType:   jschema.JSONTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							JSONType:   jschema.JSONTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							JSONType:   jschema.JSONTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							JSONType:   jschema.JSONTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							JSONType:   jschema.JSONTypeString,
							SchemaType: string(jschema.SchemaTypeString),
							Value:      "John",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"type": {
										JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							JSONType:   jschema.JSONTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							JSONType:   jschema.JSONTypeString,
							SchemaType: string(jschema.SchemaTypeString),
							Value:      "John",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"type": {
										JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							JSONType:   jschema.JSONTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							JSONType:   jschema.JSONTypeString,
							SchemaType: string(jschema.SchemaTypeString),
							Value:      "John",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"type": {
										JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							JSONType:   jschema.JSONTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeObject,
					SchemaType: string(jschema.SchemaTypeObject),
					Children: []jschema.ASTNode{
						{
							Key:        "id",
							JSONType:   jschema.JSONTypeNumber,
							SchemaType: string(jschema.SchemaTypeInteger),
							Value:      "5",
							Rules:      &jschema.RuleASTNodes{},
						},
						{
							Key:        "name",
							JSONType:   jschema.JSONTypeString,
							SchemaType: string(jschema.SchemaTypeString),
							Value:      "John",
							Rules: jschema.NewRuleASTNodes(
								map[string]jschema.RuleASTNode{
									"type": {
										JSONType:   jschema.JSONTypeString,
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
					JSONType:   jschema.JSONTypeString,
					SchemaType: string(jschema.SchemaTypeEnum),
					Value:      "foo",
					Rules: jschema.NewRuleASTNodes(
						map[string]jschema.RuleASTNode{
							"enum": {
								JSONType:   jschema.JSONTypeShortcut,
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
		}

		for given, c := range cc {
			t.Run(given, func(t *testing.T) {
				s := New("", given)

				for n, r := range c.rules {
					require.NoError(t, s.AddRule(n, enum.New(n, r)))
				}

				for n, c := range c.types {
					require.NoError(t, s.AddType(n, New(n, c)))
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
				s := New("", c.schema)

				for n, c := range c.types {
					require.NoError(t, s.AddType(n, New(n, c)))
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
				ss, err := New("", given).UsedUserTypes()
				require.NoError(t, err)
				assert.ElementsMatch(t, expected, ss)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		_, err := New("", "foo").UsedUserTypes()
		assert.EqualError(t, err, `ERROR (code 301): Invalid character "o" in literal false (expecting 'a')
	in line 1 on file 
	> foo
	---^`)
	})
}

func TestSchema_Build(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		err := New("", "42").Build()
		assert.NoError(t, err)
	})

	t.Run("negative", func(t *testing.T) {
		err := New("", "foo").Build()
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
				JSONType: jschema.JSONTypeString,
				Rules:    &jschema.RuleASTNodes{},
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
