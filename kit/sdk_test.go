package kit

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/reader"
	"github.com/jsightapi/jsight-schema-go-library/test"
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
