package loader

import "github.com/jsightapi/jsight-schema-go-library/notations/jschema/ischema"

func AddUnnamedTypes(rootSchema *ischema.ISchema) {
	for _, typ := range rootSchema.TypesList() {
		for unnamed, unnamedTyp := range typ.Schema.TypesList() {
			rootSchema.AddType(unnamed, unnamedTyp)
		}
	}
}
