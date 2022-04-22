package loader

import (
	"j/schema/fs"
	"j/schema/notations/jschema/internal/scanner"
	"j/schema/notations/jschema/internal/schema"
)

// NewSchemaForSdk reads the Schema from a file without adding to the collection.
// Does not compile allOf, in order that before there was a possibility to add
// additional TYPES.
func NewSchemaForSdk(file *fs.File, areKeysOptionalByDefault bool) *schema.Schema {
	return LoadSchema(scanner.NewSchemaScanner(file, false), nil, areKeysOptionalByDefault)
}
