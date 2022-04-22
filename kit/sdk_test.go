package kit

import (
	"fmt"
	"j/schema/fs"
	"j/schema/internal/errors"
	"j/schema/reader"
	"j/schema/test"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleLengthOfSchema() {
	// non-space characters after the top level value are allowed
	b := []byte(`
{
	"key": 123 // {min: 1} 
}
some extra text
`)

	length, err := LengthOfSchema(fs.NewFile("schema", b))

	if err != nil {
		fmt.Println("Filename:", err.Filename()) // return "schema"
		fmt.Println("Position:", err.Position())
		fmt.Println("Message:", err.Message())
		return
	}

	fmt.Println(length)
	// Output: 29
}

func ExampleLengthOfJson() {
	// non-space characters after the top level value are allowed
	b := []byte(`{ "key": 123 } some extra text`)

	length, err := LengthOfJson(fs.NewFile("json", b))

	if err != nil {
		fmt.Println("Filename:", err.Filename()) // return "json"
		fmt.Println("Position:", err.Position())
		fmt.Println("Message:", err.Message())
		return
	}

	fmt.Println(length)
	// Output: 14
}

func ExampleSchemaExample() {
	b := []byte(`
{
	"i": 123, // {min: 1}
	"s": "str",
	"b": true,
	"n": null,
	"a": [1, "str", false, null],
	"o": {
		"ii": 999 // {max: 999}
	}
}
`)

	content, err := SchemaExample(fs.NewFile("schema", b))

	if err != nil {
		fmt.Println("Filename:", err.Filename()) // return "schema"
		fmt.Println("Position:", err.Position())
		fmt.Println("Message:", err.Message())
		return
	}

	fmt.Println(string(content))
	// Output: {"i":123,"s":"str","b":true,"n":null,"a":[1,"str",false,null],"o":{"ii":999}}
}

func ExampleValidateJson() {
	schem := []byte(` {
		"aaa": 111 // {type: "@int"}
	} `)
	json := []byte(` {"aaa":333} `)

	schemaFile := fs.NewFile("schema", schem)
	jsonFile := fs.NewFile("json", json)

	// The key of extraTypes parameter is the name of the type.
	// The file name is used only for display in case of an error.
	// They may not be the same.
	extraTypes := make(map[string]*fs.File)
	extraTypes["@int"] = fs.NewFile("@int", []byte(`222 // {min: 0}`))

	err := ValidateJson(schemaFile, extraTypes, jsonFile, false)

	if err != nil {
		fmt.Println("Filename:", err.Filename())
		fmt.Println("Position:", err.Position())
		fmt.Println("Message:", err.Message())
		return
	}

	fmt.Println("wonderful json")
	// Output: wonderful json
}

func ExampleCheckSchema() {
	schem := []byte(` {
		"aaa": 111 // {type: "@int"}
	} `)
	schemaFile := fs.NewFile("schema", schem)

	// The key of extraTypes parameter is the name of the type.
	// The file name is used only for display in case of an error.
	// They may not be the same.
	extraTypes := make(map[string]*fs.File)
	extraTypes["@int"] = fs.NewFile("@int", []byte(`222 // {min: 0}`))

	err := CheckSchema(schemaFile, extraTypes)

	if err != nil {
		fmt.Println("Filename:", err.Filename())
		fmt.Println("Position:", err.Position())
		fmt.Println("Message:", err.Message())
		return
	}

	fmt.Println("wonderful schema")
	// Output: wonderful schema
}

func ExampleCheckJson() {
	json := []byte(`[ "abc", 123, true, {}, [], null ]`)

	f := fs.NewFile("json", json)
	err := CheckJson(f)

	if err != nil {
		fmt.Println("Filename:", err.Filename())
		fmt.Println("Position:", err.Position())
		fmt.Println("Message:", err.Message())
		return
	}

	fmt.Println("wonderful json")
	// Output: wonderful json
}

func TestCheckSchema(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		schemaFile := fs.NewFile("schema", []byte(`0 // {min:0}`))
		err := CheckSchema(schemaFile, nil)
		requireNoError(t, err)
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("contraint", func(t *testing.T) {
			schemaFile := fs.NewFile("schema", []byte(`-1 // {min:0}`))

			err := CheckSchema(schemaFile, nil)
			assertKitError(
				t,
				err,
				"schema",
				0,
				`Invalid value for "min" = 0 constraint `,
				errors.ErrConstraintValidation,
			)
		})

		t.Run("type not found", func(t *testing.T) {
			schemaFile := fs.NewFile("schema", []byte(`{
		"aaa": 111 // {type: "@int"}
	}`))

			err := CheckSchema(schemaFile, nil)
			assertKitError(
				t,
				err,
				"schema",
				11,
				`Type "@int" not found`,
				errors.ErrTypeNotFound,
			)
		})
	})
}

func TestValidateJson(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		schemaFile := fs.NewFile("", []byte(`
	{ // {allOf: "@aaa"}
		"bbb": 222 
	}`))

		extraTypes := map[string]*fs.File{
			"@aaa": fs.NewFile("@aaa", []byte(`{"aaa": 111}`)),
		}

		cc := []string{
			`{"aaa": 1, "bbb": 2}`,
			`{"aaa": 1}`,
			`{"bbb": 2}`,
			`{}`,
		}

		for _, json := range cc {
			t.Run(json, func(t *testing.T) {
				jsonFile := fs.NewFile("json", []byte(json))
				err := ValidateJson(schemaFile, extraTypes, jsonFile, true)
				requireNoError(t, err)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("type not found", func(t *testing.T) {
			schemaFile := fs.NewFile("schema", []byte(`{
		"aaa": 111 // {type: "@int"}
	}`))
			jsonFile := fs.NewFile("json", nil)

			err := ValidateJson(schemaFile, nil, jsonFile, false)
			assertKitError(
				t,
				err,
				"schema",
				11,
				`Type "@int" not found`,
				errors.ErrTypeNotFound,
			)
		})

		t.Run("incorrect type", func(t *testing.T) {
			schemaFile := fs.NewFile("schema", []byte(`{
		"aaa": 111 // {type: "@int"}
	}`))
			jsonFile := fs.NewFile("json", nil)

			extraTypes := map[string]*fs.File{
				"@int": fs.NewFile("@int", []byte(`"abc"`)),
			}

			err := ValidateJson(schemaFile, extraTypes, jsonFile, false)
			assertKitError(
				t,
				err,
				"schema",
				11,
				"Incorrect type of user type",
				errors.ErrIncorrectUserType,
			)
		})
	})
}

func BenchmarkValidateJson(b *testing.B) {
	testDataDir := filepath.Join(test.GetProjectRoot(), "testdata")

	jsonFile := reader.Read(filepath.Join(testDataDir, "big.json"))
	schemaFile := reader.Read(filepath.Join(testDataDir, "big.jschema"))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ValidateJson(schemaFile, nil, jsonFile, false)
	}
}

func TestSchemaExample(t *testing.T) {
	t.Run("negative", func(t *testing.T) {
		cc := map[string]string{
			"invalid type reference for type constraint": "@type",
			"invalid type reference for or constraint":   `42 // {or: ["@circle", "@square"]}`,
		}

		for name, content := range cc {
			t.Run(name, func(t *testing.T) {
				file := fs.NewFile("schema", []byte(content))
				_, err := SchemaExample(file)
				assertKitError(
					t,
					err,
					"schema",
					0,
					"Found an invalid reference to the type",
					errors.ErrUserTypeFound,
				)
			})
		}
	})
}

func TestCheckEnum(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		testList := []string{
			`[]`,
			`[1]`,
			`[1,2]`,
			`[1,2,3]`,
			`   [1,2,3]   `,
			`   [1,  2,  3]   `,
			"\n[1,2]",
			"[\n1,2]",
			"[1\n,2]",
			"[1,\n2]",
			"[1,2\n]",
			"[1,2]\n",
			`["aaa", "bbb", "ccc"]`,
			`[123, 45.67, "abc", true, false, null] `,
		}

		for _, enum := range testList {
			t.Run(enum, func(t *testing.T) {
				f := fs.NewFile("enum", []byte(enum))
				err := CheckEnum(f)
				requireNoError(t, err)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]struct {
			expectedPosition uint
			expectedMessage  string
			expectedErrCode  errors.ErrorCode
		}{
			"123": {
				expectedMessage: `An array was expected as a value for the "enum"`,
				expectedErrCode: errors.ErrArrayWasExpectedInEnumRule,
			},
			`"abc"`: {
				expectedMessage: `An array was expected as a value for the "enum"`,
				expectedErrCode: errors.ErrArrayWasExpectedInEnumRule,
			},
			"true": {
				expectedMessage: `An array was expected as a value for the "enum"`,
				expectedErrCode: errors.ErrArrayWasExpectedInEnumRule,
			},
			"false": {
				expectedMessage: `An array was expected as a value for the "enum"`,
				expectedErrCode: errors.ErrArrayWasExpectedInEnumRule,
			},
			"null": {
				expectedMessage: `An array was expected as a value for the "enum"`,
				expectedErrCode: errors.ErrArrayWasExpectedInEnumRule,
			},
			"{}": {
				expectedMessage: `An array was expected as a value for the "enum"`,
				expectedErrCode: errors.ErrArrayWasExpectedInEnumRule,
			},
			"[1,2,3] xxx": {
				expectedPosition: 8,
				expectedMessage:  "Invalid character \"x\" non-space byte after top-level value",
				expectedErrCode:  errors.ErrInvalidCharacter,
			},
			"xxx [1,2,3]": {
				expectedMessage: "Invalid character \"x\" looking for beginning of value",
				expectedErrCode: errors.ErrInvalidCharacter,
			},
			"[1,]": {
				expectedPosition: 3,
				expectedMessage:  "Invalid character \"]\" looking for beginning of value",
				expectedErrCode:  errors.ErrInvalidCharacter,
			},
			"[,1]": {
				expectedPosition: 1,
				expectedMessage:  "Invalid character \",\" looking for beginning of value",
				expectedErrCode:  errors.ErrInvalidCharacter,
			},
			"[ {} ]": {
				expectedPosition: 2,
				expectedMessage:  `Incorrect array item type in "enum". Only literals are allowed.`,
				expectedErrCode:  errors.ErrIncorrectArrayItemTypeInEnumRule,
			},
			"[ [] ]": {
				expectedPosition: 2,
				expectedMessage:  `Incorrect array item type in "enum". Only literals are allowed.`,
				expectedErrCode:  errors.ErrIncorrectArrayItemTypeInEnumRule,
			},
		}

		for enum, c := range cc {
			t.Run(enum, func(t *testing.T) {
				err := CheckEnum(fs.NewFile("enum", []byte(enum)))
				assertKitError(
					t,
					err,
					"enum",
					c.expectedPosition,
					c.expectedMessage,
					c.expectedErrCode,
				)
			})
		}
	})
}

func requireNoError(t *testing.T, err Error) {
	if assert.Nil(t, err) {
		return
	}

	t.Errorf("Got unexpected error (%d) %s", err.ErrCode(), err.Message())
}

func assertKitError(
	t *testing.T,
	err Error,
	expectedFilename string,
	expectedPosition uint,
	expectedMessage string,
	expectedCode errors.ErrorCode,
) {
	if !assert.NotNil(t, err) {
		return
	}

	assert.Equal(t, expectedFilename, err.Filename())
	assert.Equal(t, expectedPosition, err.Position())
	assert.Equal(t, expectedMessage, err.Message())
	assert.Equal(t, int(expectedCode), err.ErrCode())
}
