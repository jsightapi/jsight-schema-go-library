package checker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/internal/logger"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/loader"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/scanner"
)

func TestCheckRules(t *testing.T) {
	type typ struct {
		name   string
		schema string
	}

	cat := typ{
		"@cat",
		`{
			"catId": @catId,
			"catName": "Tom"
		}`,
	}

	catID := typ{
		"@catId",
		`12 // {min: 1}`,
	}

	check := func(schema string, types []typ) {
		schemaFile := fs.NewFile("schema", bytes.Bytes(schema))

		rootSchema := loader.NewSchemaForSdk(schemaFile, false)

		for _, typ := range types {
			f := fs.NewFile(typ.name, bytes.Bytes(typ.schema))
			ty := loader.LoadSchema(scanner.NewSchemaScanner(f, false), rootSchema, false)
			rootSchema.AddNamedType(typ.name, ty, f, 0)
		}

		loader.CompileAllOf(rootSchema)

		CheckRootSchema(rootSchema, logger.LogToNull{})
	}

	t.Run("negative", func(t *testing.T) {
		tests := []struct {
			schema string
			types  []typ
			err    errors.ErrorCode
		}{
			{
				`{
					"object": {} /* {type: "object",
							enum: ["white", "black"]}
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"object": {} /* {type: "object",
							maxItems: 10,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"object": {} /* {type: "object",
							minItems: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"object": {} /* {type: "object",
							regex: "^[A-Za-z]+$",
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"object": {} /* {type: "object",
							maxLength: 100
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"object": {} /* {type: "object",
							minLength: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"object": {} /* {type: "object",
							precision: 2,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"object": {} /* {type: "object",
							max: 1,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"object": {} /* {type: "object",
							min: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"object": {} /* {type: "object",
							or: [{type: "string"}, {type: "integer"}],
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"object": {} /* {type: "object",
							const: true
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"array" : [   /* {type: "array",
								const: true,
						}*/
						"item"
					]
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"array" : [   /* {type: "array",
								min: 0,
						}*/
						"item"
					]
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"array" : [   /* {type: "array",
								max: 1,
						}*/
						"item"
					]
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"array" : [   /* {type: "array",
								exclusiveMinimum: true,
						}*/
						"item"
					]
				}`,
				[]typ{},
				errors.ErrConstraintMinNotFound,
			},
			{
				`{
					"array" : [   /* {type: "array",
								exclusiveMaximum: true,
						}*/
						"item"
					]
				}`,
				[]typ{},
				errors.ErrConstraintMaxNotFound,
			},
			{
				`{
					"array" : [   /* {type: "array",
								precision: 2,
						}*/
						"item"
					]
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"array" : [   /* {type: "array",
								minLength: 0,
						}*/
						"item"
					]
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"array" : [   /* {type: "array",
								maxLength: 100,
						}*/
						"item"
					]
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"array" : [   /* {type: "array",
								regex: "^[A-Za-z]+$",
						}*/
						"item"
					]
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"array" : [   /* {type: "array",
								or: [{type: "string"}, {type: "integer"}],
						}*/
						"item"
					]
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"array" : [   /* {type: "array",
								additionalProperties: true,
						}*/
						"item"
					]
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"array" : [   // {allOf: "@cat"}
						"item"
					]
				}`,
				[]typ{cat, catID},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"array" : [   /* {type: "array",
								allOf: "@cat",
						}*/
						"item"
					]
				}`,
				[]typ{cat, catID},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"array" : [   /* {type: "array",
								enum: ["white", "black"]}
						}*/
						"item"
					]
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"integer": 1 /* {type: "integer",
							precision: 2,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"integer": 1 /* {type: "integer",
							minLength: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"integer": 1 /* {type: "integer",
							maxLength: 100,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"integer": 1 /* {type: "integer",
							regex: "^[A-Za-z]+$",
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"integer": 1 /* {type: "integer",
							minItems: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"integer": 1 /* {type: "integer",
							maxItems: 10,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"integer": 1 /* {type: "integer",
							or: [{type: "string"}, {type: "integer"}],
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"integer": 1 /* {type: "integer",
							additionalProperties: true,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"integer": 1 /* {type: "integer",
							allOf: "@cat",
					}*/
				}`,
				[]typ{cat, catID},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"integer": 1 /* {type: "integer",
							enum: ["white", "black"]}
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"float": 1.2 /* {type: "float",
						  precision: 1
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"float": 1.2 /* {type: "float",
						  minLength: 0
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"float": 1.2 /* {type: "float",
						  maxLength: 100
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"float": 1.2 /* {type: "float",
						  regex: "^[A-Za-z]+$"
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"float": 1.2 /* {type: "float",
						  minItems: 0
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"float": 1.2 /* {type: "float",
						  maxItems: 10
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"float": 1.2 /* {type: "float",
						  or: [{type: "string"}, {type: "float"}]
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"float": 1.2 /* {type: "float",
						  additionalProperties: true
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"float": 1.2 /* {type: "float",
						  allOf: "@cat"
					}*/
				}`,
				[]typ{cat, catID},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"float": 1.2 /* {type: "float",
						  enum: [1.2, 1.3]
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"decimal": 1.23 /* {type: "decimal", precision: 2,
							minLength: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"decimal": 1.23 /* {type: "decimal", precision: 2,
							maxLength: 100,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"decimal": 1.23 /* {type: "decimal", precision: 2,
							regex: "^[A-Za-z]+$",
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"decimal": 1.23 /* {type: "decimal", precision: 2,
							minItems: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"decimal": 1.23 /* {type: "decimal", precision: 2,
							maxItems: 10,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"decimal": 1.23 /* {type: "decimal", precision: 2,
							or: [{type: "string"}, {type: "integer"}],
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"decimal": 1.23 /* {type: "decimal", precision: 2,
							additionalProperties: true,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"decimal": 1.23 /* {type: "decimal", precision: 2,
							allOf: "@cat",
					}*/
				}`,
				[]typ{cat, catID},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"decimal": 1.23 /* {type: "decimal", precision: 2,
							enum: ["white", "black"]}
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"boolean": true /* {type: "boolean",
							min: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"boolean": true /* {type: "boolean",
							max: 1,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"boolean": true /* {type: "boolean",
							exclusiveMinimum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrConstraintMinNotFound,
			},
			{
				`{
					"boolean": true /* {type: "boolean",
							exclusiveMaximum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrConstraintMaxNotFound,
			},
			{
				`{
					"boolean": true /* {type: "boolean",
							precision: 2,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"boolean": true /* {type: "boolean",
							minLength: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"boolean": true /* {type: "boolean",
							maxLength: 100,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"boolean": true /* {type: "boolean",
							regex: "^[A-Za-z]+$",
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"boolean": true /* {type: "boolean",
							minItems: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"boolean": true /* {type: "boolean",
							maxItems: 10,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"boolean": true /* {type: "boolean",
							or: [{type: "string"}, {type: "integer"}],
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"boolean": true /* {type: "boolean",
							additionalProperties: true,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"boolean": true /* {type: "boolean",
							allOf: "@cat",
					}*/
				}`,
				[]typ{cat, catID},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"boolean": true /* {type: "boolean",
							enum: ["white", "black"]}
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"string": "value" /* {type: "string",
							min: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"string": "value" /* {type: "string",
							max: 1,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"string": "value" /* {type: "string",
							exclusiveMinimum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrConstraintMinNotFound,
			},
			{
				`{
					"string": "value" /* {type: "string",
							exclusiveMaximum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrConstraintMaxNotFound,
			},
			{
				`{
					"string": "value" /* {type: "string",
							precision: 2,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"string": "value" /* {type: "string",
							minItems: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"string": "value" /* {type: "string",
							maxItems: 10,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"string": "value" /* {type: "string",
							or: [{type: "string"}, {type: "integer"}],
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"string": "value" /* {type: "string",
							additionalProperties: true,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"string": "value" /* {type: "string",
							allOf: "@cat",
					}*/
				}`,
				[]typ{cat, catID},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"string": "value" /* {type: "string",
							enum: ["white", "black"]}
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"email": "t@t.com" /* {type: "email",
							min: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"email": "t@t.com" /* {type: "email",
							max: 1,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"email": "t@t.com" /* {type: "email",
							exclusiveMinimum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrConstraintMinNotFound,
			},
			{
				`{
					"email": "t@t.com" /* {type: "email",
							exclusiveMaximum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrConstraintMaxNotFound,
			},
			{
				`{
					"email": "t@t.com" /* {type: "email",
							precision: 2,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"email": "t@t.com" /* {type: "email",
							minLength: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"email": "t@t.com" /* {type: "email",
							maxLength: 100,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"email": "t@t.com" /* {type: "email",
							regex: ".*",
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"email": "t@t.com" /* {type: "email",
							minItems: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"email": "t@t.com" /* {type: "email",
							maxItems: 10,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uri": "https://t.com" /* {type: "uri",
							min: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uri": "https://t.com" /* {type: "uri",
							max: 1,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uri": "https://t.com" /* {type: "uri",
							exclusiveMinimum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrConstraintMinNotFound,
			},
			{
				`{
					"uri": "https://t.com" /* {type: "uri",
							exclusiveMaximum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrConstraintMaxNotFound,
			},
			{
				`{
					"uri": "https://t.com" /* {type: "uri",
							precision: 2,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uri": "https://t.com" /* {type: "uri",
							minLength: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uri": "https://t.com" /* {type: "uri",
							maxLength: 100,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uri": "https://t.com" /* {type: "uri",
							regex: ".*",
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uri": "https://t.com" /* {type: "uri",
							minItems: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uri": "https://t.com" /* {type: "uri",
							maxItems: 10,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uri": "https://t.com" /* {type: "uri",
							or: [{type: "string"}, {type: "integer"}],
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"uri": "https://t.com" /* {type: "uri",
							additionalProperties: true,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uri": "https://t.com" /* {type: "uri",
							allOf: "@cat",
					}*/
				}`,
				[]typ{cat, catID},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uri": "https://t.com" /* {type: "uri",
							enum: ["white", "black"]}
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"date": "2021-12-16" /* {type: "date",
							min: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"date": "2021-12-16" /* {type: "date",
							max: 1,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"date": "2021-12-16" /* {type: "date",
							exclusiveMinimum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrConstraintMinNotFound,
			},
			{
				`{
					"date": "2021-12-16" /* {type: "date",
							exclusiveMaximum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrConstraintMaxNotFound,
			},
			{
				`{
					"date": "2021-12-16" /* {type: "date",
							precision: 2,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"date": "2021-12-16" /* {type: "date",
							minLength: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"date": "2021-12-16" /* {type: "date",
							maxLength: 100,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"date": "2021-12-16" /* {type: "date",
							regex: ".*",
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"date": "2021-12-16" /* {type: "date",
							minItems: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"date": "2021-12-16" /* {type: "date",
							maxItems: 10,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"date": "2021-12-16" /* {type: "date",
							or: [{type: "string"}, {type: "integer"}],
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"date": "2021-12-16" /* {type: "date",
							additionalProperties: true,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"date": "2021-12-16" /* {type: "date",
							allOf: "@cat",
					}*/
				}`,
				[]typ{cat, catID},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"date": "2021-12-16" /* {type: "date",
							enum: ["white", "black"]}
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"datetime": "2006-01-02T15:04:05+07:00" /* {type: "datetime",
					   min: 0,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"datetime": "2006-01-02T15:04:05+07:00" /* {type: "datetime",
					   max: 1,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"datetime": "2006-01-02T15:04:05+07:00" /* {type: "datetime",
					   exclusiveMinimum: true,
						 }*/
				}`,
				[]typ{},
				errors.ErrConstraintMinNotFound,
			},
			{
				`{
					"datetime": "2006-01-02T15:04:05+07:00" /* {type: "datetime",
					   exclusiveMaximum: true,
						 }*/
				}`,
				[]typ{},
				errors.ErrConstraintMaxNotFound,
			},
			{
				`{
					"datetime": "2006-01-02T15:04:05+07:00" /* {type: "datetime",
					   precision: 2,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"datetime": "2006-01-02T15:04:05+07:00" /* {type: "datetime",
					   minLength: 0,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"datetime": "2006-01-02T15:04:05+07:00" /* {type: "datetime",
					   maxLength: 100,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"datetime": "2006-01-02T15:04:05+07:00" /* {type: "datetime",
					   regex: ".*",
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"datetime": "2006-01-02T15:04:05+07:00" /* {type: "datetime",
					   minItems: 0,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"datetime": "2006-01-02T15:04:05+07:00" /* {type: "datetime",
					   maxItems: 10,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"datetime": "2006-01-02T15:04:05+07:00" /* {type: "datetime",
					   or: [{type: "string"}, {type: "integer"}],
						 }*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"datetime": "2006-01-02T15:04:05+07:00" /* {type: "datetime",
					   additionalProperties: true,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"datetime": "2006-01-02T15:04:05+07:00" /* {type: "datetime",
					   allOf: "@cat",
						 }*/
				}`,
				[]typ{cat, catID},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"datetime": "2006-01-02T15:04:05+07:00" /* {type: "datetime",
					   enum: ["white", "black"]}
						 }*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"uuid": "550e8400-e29b-41d4-a716-446655440000" /* {type: "uuid",
					   min: 0,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uuid": "550e8400-e29b-41d4-a716-446655440000" /* {type: "uuid",
					   max: 1,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uuid": "550e8400-e29b-41d4-a716-446655440000" /* {type: "uuid",
					   exclusiveMinimum: true,
						 }*/
				}`,
				[]typ{},
				errors.ErrConstraintMinNotFound,
			},
			{
				`{
					"uuid": "550e8400-e29b-41d4-a716-446655440000" /* {type: "uuid",
					   exclusiveMaximum: true,
						 }*/
				}`,
				[]typ{},
				errors.ErrConstraintMaxNotFound,
			},
			{
				`{
					"uuid": "550e8400-e29b-41d4-a716-446655440000" /* {type: "uuid",
					   precision: 2,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uuid": "550e8400-e29b-41d4-a716-446655440000" /* {type: "uuid",
					   minLength: 0,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uuid": "550e8400-e29b-41d4-a716-446655440000" /* {type: "uuid",
					   maxLength: 100,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uuid": "550e8400-e29b-41d4-a716-446655440000" /* {type: "uuid",
					   regex: ".*",
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uuid": "550e8400-e29b-41d4-a716-446655440000" /* {type: "uuid",
					   minItems: 0,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uuid": "550e8400-e29b-41d4-a716-446655440000" /* {type: "uuid",
					   maxItems: 10,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uuid": "550e8400-e29b-41d4-a716-446655440000" /* {type: "uuid",
					   or: [{type: "string"}, {type: "integer"}],
						 }*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"uuid": "550e8400-e29b-41d4-a716-446655440000" /* {type: "uuid",
					   additionalProperties: true,
						 }*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uuid": "550e8400-e29b-41d4-a716-446655440000" /* {type: "uuid",
					   allOf: "@cat",
						 }*/
				}`,
				[]typ{cat, catID},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"uuid": "550e8400-e29b-41d4-a716-446655440000" /* {type: "uuid",
					   enum: ["white", "black"]}
						 }*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"enum": "white" /* {enum: ["white", "black"],
							min: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithEnum,
			},
			{
				`{
					"enum": "white" /* {enum: ["white", "black"],
							max: 1,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithEnum,
			},
			{
				`{
					"enum": "white" /* {enum: ["white", "black"],
							exclusiveMinimum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithEnum,
			},
			{
				`{
					"enum": "white" /* {enum: ["white", "black"],
							exclusiveMaximum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithEnum,
			},
			{
				`{
					"enum": "white" /* {enum: ["white", "black"],
							precision: 2,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithEnum,
			},
			{
				`{
					"enum": "white" /* {enum: ["white", "black"],
							minLength: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithEnum,
			},
			{
				`{
					"enum": "white" /* {enum: ["white", "black"],
							maxLength: 100,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithEnum,
			},
			{
				`{
					"enum": "white" /* {enum: ["white", "black"],
							regex: ".*",
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithEnum,
			},
			{
				`{
					"enum": "white" /* {enum: ["white", "black"],
							minItems: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithEnum,
			},
			{
				`{
					"enum": "white" /* {enum: ["white", "black"],
							maxItems: 10,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithEnum,
			},
			{
				`{
					"enum": "white" /* {enum: ["white", "black"],
							or: [{type: "string"}, {type: "integer"}],
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithOr,
			},
			{
				`{
					"enum": "white" /* {enum: ["white", "black"],
							additionalProperties: true,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithEnum,
			},
			{
				`{
					"enum": "white" /* {enum: ["white", "black"],
							allOf: "@cat",
					}*/
				}`,
				[]typ{cat, catID},
				errors.ErrShouldBeNoOtherRulesInSetWithEnum,
			},
			{
				`{
					"mixed": "abc" /* {or: [{type: "string"}, {type: "integer"}],
							const: true,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithOr,
			},
			{
				`{
					"mixed": "abc" /* {or: [{type: "string"}, {type: "integer"}],
							min: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithOr,
			},
			{
				`{
					"mixed": "abc" /* {or: [{type: "string"}, {type: "integer"}],
							max: 1,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithOr,
			},
			{
				`{
					"mixed": "abc" /* {or: [{type: "string"}, {type: "integer"}],
							exclusiveMinimum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithOr,
			},
			{
				`{
					"mixed": "abc" /* {or: [{type: "string"}, {type: "integer"}],
							exclusiveMaximum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithOr,
			},
			{
				`{
					"mixed": "abc" /* {or: [{type: "string"}, {type: "integer"}],
							precision: 2,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithOr,
			},
			{
				`{
					"mixed": "abc" /* {or: [{type: "string"}, {type: "integer"}],
							minLength: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithOr,
			},
			{
				`{
					"mixed": "abc" /* {or: [{type: "string"}, {type: "integer"}],
							maxLength: 100,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithOr,
			},
			{
				`{
					"mixed": "abc" /* {or: [{type: "string"}, {type: "integer"}],
							regex: ".*",
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithOr,
			},
			{
				`{
					"mixed": "abc" /* {or: [{type: "string"}, {type: "integer"}],
							minItems: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithOr,
			},
			{
				`{
					"mixed": "abc" /* {or: [{type: "string"}, {type: "integer"}],
							maxItems: 10,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithOr,
			},
			{
				`{
					"mixed": "abc" /* {or: [{type: "string"}, {type: "integer"}],
							additionalProperties: true,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithOr,
			},
			{
				`{
					"mixed": "abc" /* {or: [{type: "string"}, {type: "integer"}],
							allOf: "@cat",
					}*/
				}`,
				[]typ{cat, catID},
				errors.ErrShouldBeNoOtherRulesInSetWithOr,
			},
			{
				`{
					"mixed": "abc" /* {or: [{type: "string"}, {type: "integer"}],
							enum: ["white", "black"]}
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithOr,
			},
			{
				`{
					"any": 456 /* {type: "any",
							min: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithAny,
			},
			{
				`{
					"any": 456 /* {type: "any",
							max: 1,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithAny,
			},
			{
				`{
					"any": 456 /* {type: "any",
							exclusiveMinimum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithAny,
			},
			{
				`{
					"any": 456 /* {type: "any",
							exclusiveMaximum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithAny,
			},
			{
				`{
					"any": 456 /* {type: "any",
							precision: 2,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"any": 456 /* {type: "any",
							minLength: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithAny,
			},
			{
				`{
					"any": 456 /* {type: "any",
							maxLength: 100,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithAny,
			},
			{
				`{
					"any": 456 /* {type: "any",
							regex: ".*",
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithAny,
			},
			{
				`{
					"any": 456 /* {type: "any",
							minItems: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithAny,
			},
			{
				`{
					"any": 456 /* {type: "any",
							maxItems: 10,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithAny,
			},
			{
				`{
					"any": 456 /* {type: "any",
							or: [{type: "string"}, {type: "integer"}],
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"any": 456 /* {type: "any",
							additionalProperties: true,
					}*/
				}`,
				[]typ{},
				errors.ErrShouldBeNoOtherRulesInSetWithAny,
			},
			{
				`{
					"any": 456 /* {type: "any",
							allOf: "@cat",
					}*/
				}`,
				[]typ{cat, catID},
				errors.ErrShouldBeNoOtherRulesInSetWithAny,
			},
			{
				`{
					"any": 456 /* {type: "any",
							enum: ["white", "black"]}
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"null": null /* {type: "null",
							  min: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"null": null /* {type: "null",
							  max: 1,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"null": null /* {type: "null",
							  exclusiveMinimum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrConstraintMinNotFound,
			},
			{
				`{
					"null": null /* {type: "null",
							  exclusiveMaximum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrConstraintMaxNotFound,
			},
			{
				`{
					"null": null /* {type: "null",
							  precision: 2,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"null": null /* {type: "null",
							  minLength: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"null": null /* {type: "null",
							  maxLength: 100,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"null": null /* {type: "null",
							  regex: ".*",
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"null": null /* {type: "null",
							  minItems: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"null": null /* {type: "null",
							  maxItems: 10,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"null": null /* {type: "null",
							  or: [{type: "string"}, {type: "integer"}],
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"null": null /* {type: "null",
							  additionalProperties: true,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"null": null /* {type: "null",
							  allOf: "@cat",
					}*/
				}`,
				[]typ{cat, catID},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"null": null /* {type: "null",
							  enum: ["white", "black"]}
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"userType1": @cat  /* {
							type: "@cat",
					}*/
				}`,
				[]typ{cat, catID},
				errors.ErrInvalidChildNodeTogetherWithTypeReference,
			},
			{
				`{
					"userType1": @cat  /* {
							min: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType1": @cat  /* {
							max: 1,
					}*/
				}`,
				[]typ{},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType1": @cat  /* {
							exclusiveMinimum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType1": @cat  /* {
							exclusiveMaximum: true,
					}*/
				}`,
				[]typ{},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType1": @cat  /* {
							precision: 2,
					}*/
				}`,
				[]typ{},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"userType1": @cat  /* {
							minLength: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType1": @cat  /* {
							maxLength: 100,
					}*/
				}`,
				[]typ{},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType1": @cat  /* {
							regex: ".*",
					}*/
				}`,
				[]typ{},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType1": @cat  /* {
							minItems: 0,
					}*/
				}`,
				[]typ{},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType1": @cat  /* {
							maxItems: 10,
					}*/
				}`,
				[]typ{},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType1": @cat  /* {
							or: [{type: "string"}, {type: "integer"}],
					}*/
				}`,
				[]typ{},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType1": @cat  /* {
							additionalProperties: true,
					}*/
				}`,
				[]typ{},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType1": @cat  /* {
							allOf: "@cat",
					}*/
				}`,
				[]typ{cat, catID},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType1": @cat  /* {
							enum: ["white", "black"]}
					}*/
				}`,
				[]typ{},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"userType2": 12 /* {type: "@catId",
							const: true,
					}*/
				}`,
				[]typ{catID},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType2": 12 /* {type: "@catId",
							min: 0,
					}*/
				}`,
				[]typ{catID},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType2": 12 /* {type: "@catId",
							max: 1,
					}*/
				}`,
				[]typ{catID},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType2": 12 /* {type: "@catId",
							exclusiveMinimum: true,
					}*/
				}`,
				[]typ{catID},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType2": 12 /* {type: "@catId",
							exclusiveMaximum: true,
					}*/
				}`,
				[]typ{catID},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType2": 12 /* {type: "@catId",
							precision: 2,
					}*/
				}`,
				[]typ{catID},
				errors.ErrUnexpectedConstraint,
			},
			{
				`{
					"userType2": 12 /* {type: "@catId",
							minLength: 0,
					}*/
				}`,
				[]typ{catID},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType2": 12 /* {type: "@catId",
							maxLength: 100,
					}*/
				}`,
				[]typ{catID},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType2": 12 /* {type: "@catId",
							regex: ".*",
					}*/
				}`,
				[]typ{catID},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType2": 12 /* {type: "@catId",
							minItems: 0,
					}*/
				}`,
				[]typ{catID},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType2": 12 /* {type: "@catId",
							maxItems: 10,
					}*/
				}`,
				[]typ{catID},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType2": 12 /* {type: "@catId",
							or: [{type: "string"}, {type: "integer"}],
					}*/
				}`,
				[]typ{catID},
				errors.ErrInvalidValueInTheTypeRule,
			},
			{
				`{
					"userType2": 12 /* {type: "@catId",
							additionalProperties: true,
					}*/
				}`,
				[]typ{catID},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType2": 12 /* {type: "@catId",
							allOf: "@cat",
					}*/
				}`,
				[]typ{cat, catID},
				errors.ErrCannotSpecifyOtherRulesWithTypeReference,
			},
			{
				`{
					"userType2": 12 /* {type: "@catId",
							enum: ["white", "black"]}
					}*/
				}`,
				[]typ{catID},
				errors.ErrInvalidValueInTheTypeRule,
			},
		}

		for _, tt := range tests {
			t.Run(tt.schema, func(t *testing.T) {
				defer func() {
					r := recover()
					require.NotNil(t, r, "Panic expected")

					err, ok := r.(errors.Err)
					require.Truef(t, ok, "Unexpected error type %#v", r)

					assert.Equal(t, tt.err, err.Code())
				}()
				check(tt.schema, tt.types)
			})
		}
	})
}
