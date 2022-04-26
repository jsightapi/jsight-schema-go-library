package loader

import (
	"testing"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/scanner"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema/constraint"
)

func testSchema(areKeysOptionalByDefault bool) *schema.Schema {
	var testTypeName = "@sub"

	var testTypeSchema = `{
		"optional": 1, // {"optional": true}
		"required": 2, // {"optional": false}
		"default": 3
	}`

	var testRootSchema = `{
		"optional": 1, // {"optional": true}
		"required": 2, // {"optional": false}
		"default": 3,
		"sub-object": {
			"optional": 1, // {"optional": true}
			"required": 2, // {"optional": false}
			"default": 3
		},
		"type": @sub
	}`

	file := new(fs.File)
	file.SetContent(bytes.Bytes(testRootSchema))
	rootSchema := NewSchemaForSdk(file, areKeysOptionalByDefault) // required by default

	f := fs.NewFile("", bytes.Bytes(testTypeSchema))
	typeSchema := LoadSchema(scanner.NewSchemaScanner(f, false), rootSchema, areKeysOptionalByDefault)
	rootSchema.AddNamedType(testTypeName, typeSchema, f, 0)

	return rootSchema
}

func Test_NewSchemaForSdk_allKeysRequiredByDefault(t *testing.T) {
	rootSchema := testSchema(false)

	requiredKeysConstraint := rootSchema.RootNode().Constraint(constraint.RequiredKeysConstraintType)
	if requiredKeysConstraint == nil {
		t.Fatal("Required keys constraint not found")
	}

	keys := requiredKeysConstraint.(*constraint.RequiredKeys).Keys()
	if !(len(keys) == 4 && keys[0] == "required" && keys[1] == "default" && keys[2] == "sub-object" && keys[3] == "type") {
		t.Fatal("Incorrect required keys")
	}

	children := rootSchema.RootNode().(*schema.ObjectNode).Children()
	if len(children) != 5 {
		t.Fatal("Incorrect children length")
	}

	subObjectNode := children[3]
	requiredKeysConstraint = subObjectNode.Constraint(constraint.RequiredKeysConstraintType)
	if requiredKeysConstraint == nil {
		t.Fatal("Required keys constraint not found in sub-object")
	}

	keys = requiredKeysConstraint.(*constraint.RequiredKeys).Keys()
	if !(len(keys) == 2 && keys[0] == "required" && keys[1] == "default") {
		t.Fatal("Incorrect required keys")
	}

	requiredKeysConstraint = rootSchema.
		Type("@sub").
		RootNode().
		Constraint(constraint.RequiredKeysConstraintType) // can panic

	if requiredKeysConstraint == nil {
		t.Fatal("Required keys constraint not found in type")
	}

	keys = requiredKeysConstraint.(*constraint.RequiredKeys).Keys()
	if !(len(keys) == 2 && keys[0] == "required" && keys[1] == "default") {
		t.Fatal("Incorrect required keys in type")
	}
}

func Test_NewSchemaForSdk_allKeysOptionalByDefault(t *testing.T) {
	rootSchema := testSchema(true)

	requiredKeysConstraint := rootSchema.RootNode().Constraint(constraint.RequiredKeysConstraintType)
	if requiredKeysConstraint == nil {
		t.Fatal("Required keys constraint not found")
	}

	keys := requiredKeysConstraint.(*constraint.RequiredKeys).Keys()
	if !(len(keys) == 1 && keys[0] == "required") {
		t.Fatal("Incorrect required keys")
	}

	children := rootSchema.RootNode().(*schema.ObjectNode).Children()
	if len(children) != 5 {
		t.Fatal("Incorrect children length")
	}

	subObjectNode := children[3]
	requiredKeysConstraint = subObjectNode.Constraint(constraint.RequiredKeysConstraintType)
	if requiredKeysConstraint == nil {
		t.Fatal("Required keys constraint not found in sub-object")
	}

	keys = requiredKeysConstraint.(*constraint.RequiredKeys).Keys()
	if !(len(keys) == 1 && keys[0] == "required") {
		t.Fatal("Incorrect required keys")
	}

	requiredKeysConstraint = rootSchema.
		Type("@sub").
		RootNode().
		Constraint(constraint.RequiredKeysConstraintType) // can panic

	if requiredKeysConstraint == nil {
		t.Fatal("Required keys constraint not found in type")
	}

	keys = requiredKeysConstraint.(*constraint.RequiredKeys).Keys()
	if !(len(keys) == 1 && keys[0] == "required") {
		t.Fatal("Incorrect required keys in type")
	}
}
