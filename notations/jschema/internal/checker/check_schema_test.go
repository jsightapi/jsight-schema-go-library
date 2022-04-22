package checker

import (
	"j/schema/bytes"
	"j/schema/errors"
	"j/schema/fs"
	"j/schema/internal/logger"
	"j/schema/notations/jschema/internal/loader"
	"j/schema/notations/jschema/internal/scanner"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckRootSchema(t *testing.T) {
	type typ struct {
		name   string
		schema string
	}

	check := func(schema string, types []typ) {
		schemaFile := fs.NewFile("schema", bytes.Bytes(schema))

		rootSchema := loader.NewSchemaForSdk(schemaFile, false)

		for _, datum := range types {
			f := fs.NewFile(datum.name, bytes.Bytes(datum.schema))
			typ := loader.LoadSchema(scanner.NewSchemaScanner(f, false), rootSchema, false)
			rootSchema.AddNamedType(datum.name, typ, f, 0)
		}

		loader.CompileAllOf(rootSchema)

		CheckRootSchema(rootSchema, logger.LogToNull{})
	}

	t.Run("positive", func(t *testing.T) {
		var tests = []struct {
			schema string
			types  []typ
		}{
			{
				`{}`,
				[]typ{},
			},
			{
				`[1,2,3]`,
				[]typ{},
			},
			{
				`123`,
				[]typ{},
			},
			{
				`"qwerty"`,
				[]typ{},
			},
			{
				`{} // note`,
				[]typ{},
			},
			{
				`123 // note`,
				[]typ{},
			},
			{
				`"qwerty" // note`,
				[]typ{},
			},
			{
				"5 // {min: 1}", // the rule without quotes
				[]typ{},
			},
			{
				`5 // {"min": 1}`, // the rule with quotes
				[]typ{},
			},
			{
				`5 // {min: 1, max: 5}`, // a few rules
				[]typ{},
			},
			{
				"5 // {}", // without rules
				[]typ{},
			},
			{
				"5 // {min: 1} - some comment", // text after rule
				[]typ{},
			},
			{
				"5 // some comment", // text without rules
				[]typ{},
			},
			{
				"5 // - some comment", // text without rules
				[]typ{},
			},
			{
				"5 // [ some comment ]", // text without rules
				[]typ{},
			},

			// typeConstraint: explicit type definition
			{
				`{} // {type: "object"}`,
				[]typ{},
			},
			{
				`true // {type: "boolean"}`,
				[]typ{},
			},
			{
				`"abc" // {type: "string"}`,
				[]typ{},
			},
			{
				`null // {type: "null"}`,
				[]typ{},
			},
			{
				`[ // {type: "array"}
					1
				]`,
				[]typ{},
			},

			// precisionConstraint: decimal type
			{
				`1.1 // {precision: 1}`,
				[]typ{},
			},

			{
				`1.0 // {precision: 1}`,
				[]typ{},
			},
			{
				`1.00 // {precision: 2}`,
				[]typ{},
			},

			{
				`0.12 // {precision: 2}`,
				[]typ{},
			},

			{
				`0.120 // {precision: 2}`,
				[]typ{},
			},
			{
				`0.1200 // {precision: 2}`,
				[]typ{},
			},

			{
				`123.45 // {type: "decimal", precision: 2}`,
				[]typ{},
			},

			// min
			{
				`123 // {min: 1}`,
				[]typ{},
			},
			{
				`123 // {"min": 1}`,
				[]typ{},
			},
			{
				`-1 // {"min": -2}`,
				[]typ{},
			},
			{
				` 0 // {"min": -2}`,
				[]typ{},
			},
			{
				` 1 // {"min": -2}`,
				[]typ{},
			},

			// max
			{
				`123 // {max: 999}`,
				[]typ{},
			},
			{
				`123 // {"max": 999}`,
				[]typ{},
			},
			{
				`-1 // {"max": 1}`,
				[]typ{},
			},
			{
				` 0 // {"max": 1}`,
				[]typ{},
			},
			{
				` 1 // {"max": 1}`,
				[]typ{},
			},

			// exclusiveMinimumConstraint
			{
				`111 // {min: 1, exclusiveMinimum: true}`,
				[]typ{},
			},
			{
				`111 // {min: 1, exclusiveMinimum: false}`,
				[]typ{},
			},

			// exclusiveMaximumConstraint
			{
				`222 // {max: 333, exclusiveMaximum: true}`,
				[]typ{},
			},
			{
				`222 // {max: 333, exclusiveMaximum: false}`,
				[]typ{},
			},

			// optionalConstraints
			{
				`{
				"key": 1 // {optional: true}
			}`,
				[]typ{},
			},

			// rule "or"
			{
				`5 // {or: [ {type: "integer"}, {type: "string"} ]}`, // "or" with two simple rule-set
				[]typ{},
			},
			{
				`5 // {or: [ {min: 0}, {type: "string"} ]}`, // "or" with two simple rule-set (the first rule-set without type specifying)
				[]typ{},
			},
			{
				`5 // {or: [ {type: "@int"}, {type: "@str"} ]}`, // "or" with type names
				[]typ{
					{`@int`, `123`},
					{`@str`, `"abc"`},
				},
			},
			{
				`5 // {or: [ "@int", "@str" ]}`, // "or" with short format type names
				[]typ{
					{`@int`, `123`},
					{`@str`, `"abc"`},
				},
			},
			{
				"@int | @str", // "or" shortcut
				[]typ{
					{`@int`, `123`},
					{`@str`, `"abc"`},
				},
			},
			{
				`5 // {or: [ {type: "@int"}, "@str", {min: 0}, {type: "string"} ]}`, // "or" with different format or rule-sets
				[]typ{
					{`@int`, `123`},
					{`@str`, `"abc"`},
				},
			},

			// `{
			// 	"key": [ // {optional: true, minItems: 1}
			// 		123
			// 	]
			// }`,

			// `// text-1
			// // text-2
			//
			// { // text-3
			// 	// text-4
			// 	"aaa": 111, // {min: 1}
			// 	"bbb": 222, // {min: 2, max: 999}
			// 	"ccc": { // text-5
			// 		"ddd": 333, // {min: 3, optional: true}
			// 		"error": [] // {optional: true, minItems: 0}
			// 		// text-6
			// 	}, // text-7
			// 	"eee": 444 // {min: 4}
			// } // text-8
			//
			// // text-9
			// // text-10`,

			// some type (string)
			{
				`"abc" // {type: "@schema"}`,
				[]typ{
					{`@schema`, `"qwerty"`},
				},
			},
			{
				"@schema",
				[]typ{
					{`@schema`, `"qwerty"`},
				},
			},
			// some type (integer)
			{
				`222 // {type: "@schema"}`,
				[]typ{
					{`@schema`, `111`},
				},
			},
			{
				"@schema",
				[]typ{
					{`@schema`, `111`},
				},
			},
			// some type (float)
			{
				`3.4 // {type: "@schema"}`,
				[]typ{
					{`@schema`, `1.2`},
				},
			},
			{
				"@schema",
				[]typ{
					{`@schema`, `1.2`},
				},
			},

			// some type (boolean)
			{
				`false // {type: "@schema"}`,
				[]typ{
					{`@schema`, `true`},
				},
			},
			{
				"@schema",
				[]typ{
					{`@schema`, `true`},
				},
			},

			// some type (object)
			{
				"@schema",
				[]typ{
					{`@schema`, `{
					"key": "val"
				}`},
				},
			},

			// some type (array)
			{
				"@schema",
				[]typ{
					{`@schema`, `[1,2,3]`},
				},
			},

			// email and string
			{
				`"aaa@bbb.cc" // {type: "@email"}`,
				[]typ{
					{`@email`, `"ddd@eee.ff" // {type: "email"}`},
				},
			},
			{
				"@email",
				[]typ{
					{`@email`, `"ddd@eee.ff" // {type: "email"}`},
				},
			},

			// decimal and float
			{
				`3.4 // {type: "@schema"}`,
				[]typ{
					{`@schema`, `1.2 // {precision: 1}`},
				},
			},
			{
				"@schema",
				[]typ{
					{`@schema`, `1.2 // {precision: 1}`},
				},
			},

			// or
			{
				`222 // {or: ["@int","@str"]}`,
				[]typ{
					{`@int`, `111`},
					{`@str`, `"abc"`},
				},
			},
			{
				`"str" // {or: ["@int","@str"]}`,
				[]typ{
					{`@int`, `111`},
					{`@str`, `"abc"`},
				},
			},
			{
				`222 // {or: [ {type:"@int"}, {type:"@str"} ]}`,
				[]typ{
					{`@int`, `111`},
					{`@str`, `"abc"`},
				},
			},
			{
				`"str" // {or: [ {type:"@int"}, {type:"@str"} ]}`,
				[]typ{
					{`@int`, `111`},
					{`@str`, `"abc"`},
				},
			},

			{
				`  222 // {or: [ {type:"integer"}, {type:"string"} ]}`,
				[]typ{},
			},
			{
				`"str" // {or: [ {type:"integer"}, {type:"string"} ]}`,
				[]typ{},
			},

			{
				`1 // {or: ["@int_or_str", "@obj"]}`,
				[]typ{
					{`@int_or_str`, `"abc" // {or: ["@int", "@str"]}`},
					{`@str`, `"abc"`},
					{`@int`, `123`},
					{`@obj`, `{}`},
				},
			},

			{
				`{
				"id": 1,
				"children": [
					@node
				]
			}`,
				[]typ{
					{"@node", `{
					"id": 1,
					"children": [
						@node
					]
				}`},
				},
			},

			{
				`1 // {type: "@type1"}`,
				[]typ{
					{`@type1`, `1 // {or: [ {type:"integer"}, {type:"string"} ]}`},
				},
			},

			// allowed recursions
			{
				"@type1",
				[]typ{
					{"@type1", "@type2"},
					{"@type2", "{}"},
				},
			},

			{
				"@user",
				[]typ{
					{`@user`, `{
					"name": "John",
					"best_friend": @user
				}`},
				},
			},

			// array and its required element
			{
				`[1]`,
				[]typ{},
			},
			{
				`[1,2]`,
				[]typ{},
			},
			{
				"@arr",
				[]typ{
					{`@arr`, `[1,2,3]`},
				},
			},
			{
				"@arr-1",
				[]typ{
					{"@arr-1", "@arr-2"},
					{"@arr-2", "[1,2,3]"},
				},
			},

			// Allow incorrect links-type for unused types
			{
				`1 // {type: "@used"}`,
				[]typ{
					{`@used`, `111`},
					{`@unused-1`, `222 // {type: "@unused-2"}`},
					{`@unused-2`, `333`},
				},
			},

			{
				`"abc" // {type: "enum", enum: [123, "abc"]}`,
				[]typ{},
			},
			{
				`"abc" // {type: "mixed", or: [{type:"integer"}, {type:"string"}]}`,
				[]typ{},
			},

			{
				`{
				"key": "abc" // {type: "mixed", or: [{type:"string"}, {type:"integer"}], optional: true}
			}`,
				[]typ{},
			},
			{
				`{
				"key": "abc" // {type: "enum", enum: [123, "abc"], optional: true}
			}`,
				[]typ{},
			},
			{
				`123   // {type: "any"}`,
				[]typ{},
			},
			{
				`12.3  // {type: "any"}`,
				[]typ{},
			},
			{
				`"str" // {type: "any"}`,
				[]typ{},
			},
			{
				`true  // {type: "any"}`,
				[]typ{},
			},
			{
				`false // {type: "any"}`,
				[]typ{},
			},
			{
				`null  // {type: "any"}`,
				[]typ{},
			},
			{
				`{}    // {type: "any"}`,
				[]typ{},
			},
			{
				`[]    // {type: "any"}`,
				[]typ{},
			},
			{
				`{
				"aaa": 1 // {type: "any", optional: true}
			}`,
				[]typ{},
			},

			{
				`[ // {minItems: 2}
				1
			]`,
				[]typ{},
			},
			{
				`[ // {maxItems: 2}
				1,2,3
			]`,
				[]typ{},
			},
			{
				`[]`,
				[]typ{},
			},
			{
				`[] // {minItems: 0}`,
				[]typ{},
			},
			{
				`[] // {maxItems: 0}`,
				[]typ{},
			},
			{
				`[] // {minItems: 0, maxItems: 0}`,
				[]typ{},
			},
			{
				`[] // {type: "array"}`,
				[]typ{},
			},
			{
				"@arr",
				[]typ{
					{"@arr", "[]"},
				},
			},

			{
				"@sub",
				[]typ{
					{"@sub", `{"_id": 123}`},
				},
			},

			{
				"@sub9",
				[]typ{
					{"@sub1", `{"_id\"": 123}`},
					{"@sub2", `{"_id\\": 123}`},
					{"@sub3", `{"_id\/": 123}`},
					{"@sub4", `{"_id\b": 123}`},
					{"@sub5", `{"_id\f": 123}`},
					{"@sub6", `{"_id\n": 123}`},
					{"@sub7", `{"_id\r": 123}`},
					{"@sub8", `{"_id\t": 123}`},
					{"@sub9", `{"_id\uAAAA": 123}`},
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.schema, func(t *testing.T) {
				check(tt.schema, tt.types)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		tests := []struct {
			schema string
			types  []typ
			err    errors.ErrorCode
		}{
			// min
			{`-1 // {"min": 2}`, []typ{}, errors.ErrConstraintValidation},
			{` 0 // {"min": 2}`, []typ{}, errors.ErrConstraintValidation},
			{` 1 // {"min": 2}`, []typ{}, errors.ErrConstraintValidation},
			{`-3 // {"min": -2}`, []typ{}, errors.ErrConstraintValidation},
			{`-4 // {"min": -2}`, []typ{}, errors.ErrConstraintValidation},

			{` 3 // {"max": 2}`, []typ{}, errors.ErrConstraintValidation},
			{` 4 // {"max": 2}`, []typ{}, errors.ErrConstraintValidation},
			{`-1 // {"max": -2}`, []typ{}, errors.ErrConstraintValidation},
			{` 0 // {"max": -2}`, []typ{}, errors.ErrConstraintValidation},
			{` 1 // {"max": -2}`, []typ{}, errors.ErrConstraintValidation},

			{"2 // {unknown: 123}", []typ{}, errors.ErrUnknownRule},
			{`2 // {type: "unknown"}`, []typ{}, errors.ErrUnknownType},

			{"2 // {min: [1,2]}", []typ{}, errors.ErrIncorrectRuleValueType},
			{"2 // {min: {}}", []typ{}, errors.ErrIncorrectRuleValueType},

			{`222 // {min: 1, min: 1}`, []typ{}, errors.ErrDuplicateRule}, // duplicate, with the same values
			{`333 // {min: 1, min: 2}`, []typ{}, errors.ErrDuplicateRule}, // duplicate, with different values

			{`{key: 1}`, []typ{}, errors.ErrInvalidCharacter},            // incorrect json example (no quotes)
			{`{"k":1, "k":2}`, []typ{}, errors.ErrDuplicateKeysInSchema}, // duplicate keys on JSON

			{`{"key": "value"} // {min: 1}`, []typ{}, errors.ErrIncorrectRuleForSeveralNode},
			{`[
			1,2 // {min: 1}
		]`, []typ{}, errors.ErrIncorrectRuleForSeveralNode},
			{`{"key": {"bbb": // {min: 1}
			2
		}}`, []typ{}, errors.ErrIncorrectRuleForSeveralNode},
			{`[
			1
		] // {min: 1}`, []typ{}, errors.ErrAnnotationNotAllowed},

			{`{
		} // {min: 1}`, []typ{}, errors.ErrIncorrectRuleWithoutExample},

			// error on constraint validation
			{`3 // {min: 4}`, []typ{}, errors.ErrConstraintValidation},
			{`3 // {max: 2}`, []typ{}, errors.ErrConstraintValidation},
			{`"str" // {minLength: 4}`, []typ{}, errors.ErrConstraintStringLengthValidation},
			{`"str" // {maxLength: 2}`, []typ{}, errors.ErrConstraintStringLengthValidation},

			// Incompatible JSON-type
			{`[ // {min: 1}
			1
		]`, []typ{}, errors.ErrUnexpectedConstraint},
			{`1 // {minLength: 1}`, []typ{}, errors.ErrUnexpectedConstraint},
			{`"str" // {min: 1}`, []typ{}, errors.ErrUnexpectedConstraint},

			// Unable to add constraint "email" to the node "Integer"
			{`123 // {type: "email"}`, []typ{}, errors.ErrUnexpectedConstraint},

			// email
			{`"" // {type: "email"}`, []typ{}, errors.ErrEmptyEmail},
			{`"no email" // {type: "email"}`, []typ{}, errors.ErrInvalidEmail},

			// Invalid value of constraint
			{`1 // {min: 99, exclusiveMinimum: 1}`, []typ{}, errors.ErrInvalidValueOfConstraint},
			{`1 // {max: 99, exclusiveMaximum: 1}`, []typ{}, errors.ErrInvalidValueOfConstraint},
			{`1.1 // {precision: true}`, []typ{}, errors.ErrInvalidValueOfConstraint},
			{`{
					"k": 1 // {optional: 1}
		}`, []typ{}, errors.ErrInvalidValueOfConstraint},

			// typeConstraint: incorrect type conversion
			{`"abc" // {type: "integer"}`, []typ{}, errors.ErrIncompatibleTypes},
			{`12.34 // {type: "integer"}`, []typ{}, errors.ErrIncompatibleTypes},
			{`123 // {type: "string"}`, []typ{}, errors.ErrIncompatibleTypes},
			{`true // {type: "string"}`, []typ{}, errors.ErrIncompatibleTypes},
			{`null // {type: "string"}`, []typ{}, errors.ErrIncompatibleTypes},
			{`{} // {type: "string"}`, []typ{}, errors.ErrIncompatibleTypes},
			{`[] // {type: "string"}`, []typ{}, errors.ErrIncompatibleTypes},
			{`123 // {type: "float"}`, []typ{}, errors.ErrIncompatibleTypes},

			// precisionConstraint
			{`123.45 // {type: "decimal"}`, []typ{}, errors.ErrNotFoundRulePrecision},          // decimal without precision
			{`123 // {precision: 1}`, []typ{}, errors.ErrUnexpectedConstraint},                 // incorrect integer node type
			{`"str" // {precision: 2}`, []typ{}, errors.ErrUnexpectedConstraint},               // incorrect string node type
			{`true // {precision: 2}`, []typ{}, errors.ErrUnexpectedConstraint},                // incorrect bool node type
			{`null // {precision: 2}`, []typ{}, errors.ErrUnexpectedConstraint},                // incorrect null node type
			{`"str" // {minLength: 0, precision: 1}`, []typ{}, errors.ErrUnexpectedConstraint}, // incompatibility node type with constraint
			{`1.0 // {precision: 0}`, []typ{}, errors.ErrZeroPrecision},                        // zero precision
			{`0.12 // {precision: -2}`, []typ{}, errors.ErrInvalidValueOfConstraint},           // negative precision
			{`0.12 // {precision: 2.3}`, []typ{}, errors.ErrInvalidValueOfConstraint},          // fractional precision

			// exclusiveMinimumConstraint
			{`111 // {exclusiveMinimum: true}`, []typ{}, errors.ErrConstraintMinNotFound},
			{`111 // {min: 2, exclusiveMinimum: 1}`, []typ{}, errors.ErrInvalidValueOfConstraint}, // not bool in exclusive

			// exclusiveMaximumConstraint
			{`222 // {exclusiveMaximum: true}`, []typ{}, errors.ErrConstraintMaxNotFound},
			{`222 // {max: 2, exclusiveMaximum: "str"}`, []typ{}, errors.ErrInvalidValueOfConstraint}, // not bool in exclusive

			// optionalConstraints: Incorrect rule "optional" location. The rule "optional" applies only to object properties.
			{`"str" // {optional: true}`, []typ{}, errors.ErrRuleOptionalAppliesOnlyToObjectProperties},
			{`12 // {optional: true}`, []typ{}, errors.ErrRuleOptionalAppliesOnlyToObjectProperties},
			{`1.2 // {optional: true}`, []typ{}, errors.ErrRuleOptionalAppliesOnlyToObjectProperties},
			{`true // {optional: true}`, []typ{}, errors.ErrRuleOptionalAppliesOnlyToObjectProperties},
			{`null // {optional: true}`, []typ{}, errors.ErrRuleOptionalAppliesOnlyToObjectProperties},
			{`{} // {optional: true}`, []typ{}, errors.ErrRuleOptionalAppliesOnlyToObjectProperties},
			{`[] // {optional: true}`, []typ{}, errors.ErrRuleOptionalAppliesOnlyToObjectProperties},
			{`[
				1 // {optional: true}
			]`, []typ{}, errors.ErrRuleOptionalAppliesOnlyToObjectProperties},
			{`{ // {optional: true}
            }`, []typ{}, errors.ErrRuleOptionalAppliesOnlyToObjectProperties},

			// You cannot specify children node if you use a type reference.
			{`{ // {type: "@schema"}
			"key": 123
		}`, []typ{}, errors.ErrInvalidChildNodeTogetherWithTypeReference},

			// You cannot specify other rules if you use a type reference.
			{`333 // {type: "@type", min: 1}`, []typ{}, errors.ErrCannotSpecifyOtherRulesWithTypeReference},
			{`333 // {type: "@type", min: 1}`, []typ{}, errors.ErrCannotSpecifyOtherRulesWithTypeReference},
			{`333 // {type: "@type1", type: "@type2"}`, []typ{}, errors.ErrDuplicateRule},

			// rule "or"
			{`2 // {or: 123}`, []typ{}, errors.ErrArrayWasExpectedInOrRule},
			{`2 // {or: "some_string"}`, []typ{}, errors.ErrArrayWasExpectedInOrRule},
			{`2 // {or: "@some_string"}`, []typ{}, errors.ErrArrayWasExpectedInOrRule},
			{`2 // {or: true}`, []typ{}, errors.ErrArrayWasExpectedInOrRule},
			{`2 // {or: null}`, []typ{}, errors.ErrArrayWasExpectedInOrRule},
			{`2 // {or: {}`, []typ{}, errors.ErrArrayWasExpectedInOrRule},
			{`2 // {or: {or: {"@type-1","@type-2"}`, []typ{}, errors.ErrArrayWasExpectedInOrRule},

			{`2 // {or: [ 1,2,3 ]}`, []typ{}, errors.ErrIncorrectArrayItemTypeInOrRule},
			{`2 // {or: [ [],[] ]}`, []typ{}, errors.ErrIncorrectArrayItemTypeInOrRule},

			{`2 // {or: [ {type: false}, {type: "string"} ]}`, []typ{}, errors.ErrUnknownType},
			{`2 // {or: [ {type: "unknown_json_type"}, {type: "string"} ]}`, []typ{}, errors.ErrUnknownType},

			{`2 // {or: [ {type: "@type", min: 0}, {type: "string"} ]}`, []typ{}, errors.ErrCannotSpecifyOtherRulesWithTypeReference},
			{`2 // {or: [ {min: 0, type: "@type"}, {type: "string"} ]}`, []typ{}, errors.ErrCannotSpecifyOtherRulesWithTypeReference},

			{`2 // {or: [ {}, {} ]}`, []typ{}, errors.ErrEmptyRuleSet},
			{`2 // {or: []}`, []typ{}, errors.ErrEmptyArrayInOrRule},

			{`2 // {or: [ {type: "integer", min: 0} ]}`, []typ{}, errors.ErrOneElementInArrayInOrRule},
			{`2 // {or: [ {min: 0} ]}`, []typ{}, errors.ErrOneElementInArrayInOrRule},
			{`2 // {or: [ {type: "@type"} ]}`, []typ{}, errors.ErrOneElementInArrayInOrRule},
			{`2 // {or: [ "@type" ]}`, []typ{}, errors.ErrOneElementInArrayInOrRule},

			//{`2 // {or: [ {type: "integer"}, {minLength:1} ]}`, []typ{}, errors.ErrIncompatibleJsonType},
			//{`2 // {or: [ {type: "integer"}, {min:1, minLength:1} ]}`, []typ{}, errors.ErrIncompatibleJsonType},

			{`2 // {or: [ "some_string" ]}`, []typ{}, errors.ErrUnknownType},
			{`2 // {or: [ {type: "integer", min: 0, min: 0}, {type: "string"} ]}`, []typ{}, errors.ErrDuplicateRule},
			{`2 // {or: [ {min: []}, {type: "string"} ]}`, []typ{}, errors.ErrLiteralValueExpected},

			{`2 // {min: 1, or: [ {type: "integer"}, {type: "string"} ]}`, []typ{}, errors.ErrShouldBeNoOtherRulesInSetWithOr},
			{
				`{ // {or: [ {type: "object"}, {type: "string"} ]}
				"key": 1
			}`,
				[]typ{},
				errors.ErrInvalidChildNodeTogetherWithOrRule,
			},
			{
				`[ // {or: [ {type: "array"}, {type: "string"} ]}
				1,2,3
			]`,
				[]typ{},
				errors.ErrInvalidChildNodeTogetherWithOrRule,
			},

			{
				``,
				[]typ{
					{`abc`, `{}`},
				},
				errors.ErrInvalidSchemaName,
			},

			{`-5 // {or: [ {min: 0}, {type: "string"} ]}`, []typ{}, errors.ErrOrRuleSetValidation},

			// duplicate type names
			{
				``,
				[]typ{
					{`@sub1`, `"some string 1"`},
					{`@sub1`, `{}`},
				},
				errors.ErrDuplicationOfNameOfTypes,
			},

			// invalid place for comment
			{
				``,
				[]typ{
					{`@sub`, `3.4
					// {precision: 1}`},
				},
				errors.ErrIncorrectRuleWithoutExample,
			},

			// schema and example type mismatch
			{
				`[] // {or: [ {type:"string"}, {type:"@arr"} ]}`,
				[]typ{
					{`@arr`, `[1,2,3]`},
				},
				errors.ErrInvalidChildNodeTogetherWithOrRule,
			},
			{
				`[] // {type: "@schema"}`,
				[]typ{
					{`@schema`, `{}`},
				},
				errors.ErrInvalidChildNodeTogetherWithTypeReference,
			},
			{
				`{} // {type: "@schema"}`,
				[]typ{
					{`@schema`, `[1]`},
				},
				errors.ErrInvalidChildNodeTogetherWithTypeReference,
			},
			{
				`"" // {type: "@schema"}`,
				[]typ{
					{`@schema`, `[1]`},
				},
				errors.ErrIncorrectUserType,
			},
			{
				`11 // {type: "@schema"}`,
				[]typ{
					{`@schema`, `[1]`},
				},
				errors.ErrIncorrectUserType,
			},
			{
				`1.2 // {type: "@schema"}`,
				[]typ{
					{`@schema`, `[1]`},
				},
				errors.ErrIncorrectUserType,
			},
			{
				`true // {type: "@schema"}`,
				[]typ{
					{`@schema`, `[1]`},
				},
				errors.ErrIncorrectUserType,
			},
			{
				`null // {type: "@schema"}`,
				[]typ{
					{`@schema`, `[1]`},
				},
				errors.ErrIncorrectUserType,
			},
			{
				`111 // {type: "@schema"}`,
				[]typ{
					{`@schema`, `1.2`},
				},
				errors.ErrIncorrectUserType,
			},
			{ // decimal and integer
				`3 // {type: "@schema"}`,
				[]typ{
					{`@schema`, `1.2 // {precision: 1}`},
				},
				errors.ErrIncorrectUserType,
			},

			// or
			{
				`1.2  // {or: ["@int","@str"]}`,
				[]typ{
					{`@int`, `111`},
					{`@str`, `"abc"`},
				},
				errors.ErrIncorrectUserType,
			},
			{
				`true // {or: ["@int","@str"]}`,
				[]typ{
					{`@int`, `111`},
					{`@str`, `"abc"`},
				},
				errors.ErrIncorrectUserType},
			{
				`null // {or: ["@int","@str"]}`,
				[]typ{
					{`@int`, `111`},
					{`@str`, `"abc"`},
				},
				errors.ErrIncorrectUserType,
			},
			{
				`{}   // {or: ["@int","@str"]}`,
				[]typ{
					{`@int`, `111`},
					{`@str`, `"abc"`},
				},
				errors.ErrInvalidChildNodeTogetherWithOrRule,
			},
			{
				`[]   // {or: ["@int","@str"]}`,
				[]typ{
					{`@int`, `111`},
					{`@str`, `"abc"`},
				},
				errors.ErrInvalidChildNodeTogetherWithOrRule,
			},

			{
				`1.2  // {or: [ {type:"@int"}, {type:"@str"} ]}`,
				[]typ{
					{`@int`, `111`},
					{`@str`, `"abc"`},
				},
				errors.ErrIncorrectUserType},
			{
				`true // {or: [ {type:"@int"}, {type:"@str"} ]}`,
				[]typ{
					{`@int`, `111`},
					{`@str`, `"abc"`},
				},
				errors.ErrIncorrectUserType},
			{
				`null // {or: [ {type:"@int"}, {type:"@str"} ]}`,
				[]typ{
					{`@int`, `111`},
					{`@str`, `"abc"`},
				},
				errors.ErrIncorrectUserType},
			{
				`{}   // {or: [ {type:"@int"}, {type:"@str"} ]}`,
				[]typ{
					{`@int`, `111`},
					{`@str`, `"abc"`},
				},
				errors.ErrInvalidChildNodeTogetherWithOrRule,
			},
			{
				`[]   // {or: [ {type:"@int"}, {type:"@str"} ]}`,
				[]typ{
					{`@int`, `111`},
					{`@str`, `"abc"`},
				},
				errors.ErrInvalidChildNodeTogetherWithOrRule,
			},

			{`1.2  // {or: [ {type:"integer"}, {type:"string"} ]}`, []typ{}, errors.ErrIncorrectUserType},
			{`true // {or: [ {type:"integer"}, {type:"string"} ]}`, []typ{}, errors.ErrIncorrectUserType},
			{`null // {or: [ {type:"integer"}, {type:"string"} ]}`, []typ{}, errors.ErrIncorrectUserType},
			{`{}   // {or: [ {type:"integer"}, {type:"string"} ]}`, []typ{}, errors.ErrIncorrectUserType},
			{`[]   // {or: [ {type:"integer"}, {type:"string"} ]}`, []typ{}, errors.ErrIncorrectUserType},

			{
				`false // {or: ["@int_or_str", "@obj"]}`,
				[]typ{
					{`@int_or_str`, `"abc" // {or: ["@int", "@str"]}`},
					{`@str`, `"abc"`},
					{`@int`, `123`},
					{`@obj`, `{}`},
				},
				errors.ErrIncorrectUserType},

			{
				`1 // {type: "@type1"}`,
				[]typ{
					{`@type1`, `1 // {type: "@type1"}`},
				},
				errors.ErrImpossibleToDetermineTheJsonTypeDueToRecursion},

			{
				`1 // {type: "@type1"}`,
				[]typ{
					{`@type1`, `1 // {type: "@type2"}`},
					{`@type2`, `2 // {type: "@type1"}`},
				},
				errors.ErrImpossibleToDetermineTheJsonTypeDueToRecursion},

			{
				`1 // {type: "@type1"}`,
				[]typ{
					{`@type1`, `1 // {type: "@type2"}`},
					{`@type2`, `2 // {type: "@type3"}`},
					{`@type3`, `3 // {type: "@type1"}`},
				},
				errors.ErrImpossibleToDetermineTheJsonTypeDueToRecursion},

			{
				`1 // {type: "@type1"}`,
				[]typ{
					{`@type1`, `1 // {or: [ {type:"integer"}, "@type2" ]}`},
					{`@type2`, `2 // {or: [ {type:"integer"}, "@type1" ]}`},
				},
				errors.ErrImpossibleToDetermineTheJsonTypeDueToRecursion},

			{
				`1 // {type: "@recurring"}`,
				[]typ{
					{`@recurring`, `"abc" // {or: ["@int", "@recurring"]}`},
					{`@int`, `123`},
				},
				errors.ErrImpossibleToDetermineTheJsonTypeDueToRecursion},

			{`"abc" // {type: "enum"}`, []typ{}, errors.ErrNotFoundRuleEnum},
			{`"abc" // {type: "enum", minLength: 1}`, []typ{}, errors.ErrNotFoundRuleEnum},

			{`"abc" // {type: "mixed"}`, []typ{}, errors.ErrNotFoundRuleOr},
			{`"abc" // {type: "mixed", minLength: 1}`, []typ{}, errors.ErrNotFoundRuleOr},

			{`2.0 // {enum: [2]}`, []typ{}, errors.ErrDoesNotMatchAnyOfTheEnumValues},
			{`2 // {enum: [2.0]}`, []typ{}, errors.ErrDoesNotMatchAnyOfTheEnumValues},

			{`"abc" // {type: "string", enum: [123, "abc"]}`, []typ{}, errors.ErrInvalidValueInTheTypeRule},
			{`"abc" // {type: "integer", enum: [123, "abc"]}`, []typ{}, errors.ErrInvalidValueInTheTypeRule},
			{`"abc" // {type: "boolean", enum: [123, "abc"]}`, []typ{}, errors.ErrInvalidValueInTheTypeRule},
			{`"abc" // {type: "string", or: [{type:"integer"}, {type:"string"}]}`, []typ{}, errors.ErrInvalidValueInTheTypeRule},

			{`2 // {type: "integer", or: [ {type: "integer"}, {type: "string"} ]}`, []typ{}, errors.ErrInvalidValueInTheTypeRule},
			{`2 // {type: "@type", or: [ {type: "integer"}, {type: "string"} ]}`, []typ{}, errors.ErrInvalidValueInTheTypeRule},

			{`"abc" // {enum: [123, "abc"], minLength: 1}`, []typ{}, errors.ErrShouldBeNoOtherRulesInSetWithEnum},
			{`"abc" // {type: "enum", enum: [123, "abc"], min: 1}`, []typ{}, errors.ErrShouldBeNoOtherRulesInSetWithEnum},

			{`123 // {type: "any", min: 1}`, []typ{}, errors.ErrShouldBeNoOtherRulesInSetWithAny},
			{`123 // {min: 1, type: "any"}`, []typ{}, errors.ErrShouldBeNoOtherRulesInSetWithAny},
			{`{ // {type: "any"}
			"aaa": 1,
			"bbb": 2
		}`, []typ{}, errors.ErrInvalidNestedElementsFoundForTypeAny},
			{`[ // {type: "any"}
			1,2,3
		]`, []typ{}, errors.ErrInvalidNestedElementsFoundForTypeAny},

			{`1 // {type: "@int"}`,
				[]typ{
					{`@int`, `1 // {type: "@str"}`},
					{`@str`, `"abc"`},
				},
				errors.ErrIncorrectUserType},
			{`1 // {type: "@int"}`,
				[]typ{
					{`@int`, `"abc" // {type: "@str"}`},
					{`@str`, `"abc"`},
				},
				errors.ErrIncorrectUserType},

			{`"abc"`,
				[]typ{
					{`@unused`, `-1 // {min: 0}`},
				},
				errors.ErrConstraintValidation},

			{`-5 // {or: [ {min: 0}, {type: "string"}, "@used" ]}`,
				[]typ{
					{`@used`, `0 // {min: -10}`},
					{`@unused`, `-1 // {min: 0} - incorrect EXAMPLE value`},
				},
				errors.ErrConstraintValidation},

			{`-1 // {type: "@int"}`,
				[]typ{
					{`@int`, `0 // {min: 0}`},
				},
				errors.ErrConstraintValidation},

			{`-1 // {type: "@int"}`,
				[]typ{
					{`@int`, `0 // {type: "@uint"}`},
					{`@uint`, `0  // {min: 0}`},
				},
				errors.ErrConstraintValidation},

			{`-1 // {type: "@int"}`,
				[]typ{
					{`@int`, `-1 // {type: "@uint"}`},
					{`@uint`, `0  // {min: 0}`},
				},
				errors.ErrConstraintValidation},

			{`{} // {additionalProperties: "wrong"}`, []typ{}, errors.ErrUnknownJSchemaType},

			{`"abc" // {additionalProperties: "string"}`, []typ{}, errors.ErrUnexpectedConstraint},
			{`123 // {additionalProperties: "string"}`, []typ{}, errors.ErrUnexpectedConstraint},
			{`123.45 // {additionalProperties: "string"}`, []typ{}, errors.ErrUnexpectedConstraint},
			{`true // {additionalProperties: "string"}`, []typ{}, errors.ErrUnexpectedConstraint},
			{`false // {additionalProperties: "string"}`, []typ{}, errors.ErrUnexpectedConstraint},
			{`null // {additionalProperties: "string"}`, []typ{}, errors.ErrUnexpectedConstraint},
			{`[ // {additionalProperties: "string"}
			123
		]`, []typ{}, errors.ErrUnexpectedConstraint},

			{`{} // {allOf: 123}`, []typ{}, errors.ErrUnacceptableValueInAllOfRule},
			{`{} // {allOf: true}`, []typ{}, errors.ErrUnacceptableValueInAllOfRule},
			{`{} // {allOf: false}`, []typ{}, errors.ErrUnacceptableValueInAllOfRule},
			{`{} // {allOf: null}`, []typ{}, errors.ErrUnacceptableValueInAllOfRule},
			{`{} // {allOf: {}}`, []typ{}, errors.ErrUnacceptableValueInAllOfRule},
			{`{} // {allOf: []}`, []typ{}, errors.ErrTypeNameNotFoundInAllOfRule},
			{`{} // {allOf: "not a schema name"}`, []typ{}, errors.ErrInvalidSchemaNameInAllOfRule},
			{`{} // {allOf: ["not a schema name"]}`, []typ{}, errors.ErrInvalidSchemaNameInAllOfRule},
			{
				`{ // {allOf: "@basicError"}
				"message": "Some message text"
			}`,
				[]typ{
					{`@basicError`, `{"message": "Some message text"}`},
				},
				errors.ErrDuplicateKeysInSchema,
			},
			{
				``,
				[]typ{
					{`@aaa`, `{ // {allOf: "@bbb"}
						"aaa": "aaa"
					}`},
					{`@bbb`, `{ // {allOf: "@aaa"}
						"bbb": "bbb"
					}`},
				},
				errors.ErrUnacceptableRecursionInAllOfRule,
			},
			{
				`{} // {allOf: "@aaa"}`,
				[]typ{
					{`@aaa`, `{ // {allOf: "@bbb"}
						"aaa": "aaa"
					}`},
					{`@bbb`, `{ // {allOf: "@aaa"}
						"bbb": "bbb"
					}`},
				},
				errors.ErrUnacceptableRecursionInAllOfRule,
			},
			{
				``,
				[]typ{
					{`@aaa`, `{ // {allOf: "@bbb"}
						"aaa": "aaa"
					}`},
				},
				errors.ErrTypeNotFound,
			},
			{
				`{} // {allOf: "@aaa"}`,
				[]typ{},
				errors.ErrTypeNotFound,
			},
			{
				`{} // {allOf: "@aaa"}`,
				[]typ{
					{`@aaa`, `[]`},
				},
				errors.ErrUnacceptableUserTypeInAllOfRule,
			},
			{
				`{} // {allOf: "@aaa"}`,
				[]typ{
					{`@aaa`, `"string"`},
				},
				errors.ErrUnacceptableUserTypeInAllOfRule,
			},
			{
				`{} // {allOf: "@aaa"}`,
				[]typ{
					{`@aaa`, `123`},
				},
				errors.ErrUnacceptableUserTypeInAllOfRule,
			},
			{
				`{} // {allOf: "@aaa"}`,
				[]typ{
					{`@aaa`, `123.45`},
				},
				errors.ErrUnacceptableUserTypeInAllOfRule,
			},
			{
				`{} // {allOf: "@aaa"}`,
				[]typ{
					{`@aaa`, `true`},
				},
				errors.ErrUnacceptableUserTypeInAllOfRule,
			},
			{
				`{} // {allOf: "@aaa"}`,
				[]typ{
					{`@aaa`, `null`},
				},
				errors.ErrUnacceptableUserTypeInAllOfRule,
			},
			{
				`{ // {allOf: "@aaa", additionalProperties: "integer"}
				"bbb": 222
			}`,
				[]typ{
					{`@aaa`, `{ // {additionalProperties: "string"}
						"aaa": 111
					}`},
				},
				errors.ErrConflictAdditionalProperties,
			},
			{
				`{ // {allOf: "@aaa", additionalProperties: "@int"}
				"bbb": 222
			}`,
				[]typ{
					{`@aaa`, `{ // {additionalProperties: "@str"}
						"aaa": 111
					}`},
					{`@str`, `"abc"`},
					{`@int`, `123`},
				},
				errors.ErrConflictAdditionalProperties,
			},
			{
				`[] // {minItems: 1}`,
				[]typ{},
				errors.ErrIncorrectConstraintValueForEmptyArray,
			},
			{
				`[] // {maxItems: 1}`,
				[]typ{},
				errors.ErrIncorrectConstraintValueForEmptyArray,
			},
			{
				`[] // {minItems: 1, maxItems: 1}`,
				[]typ{},
				errors.ErrIncorrectConstraintValueForEmptyArray,
			},
			{
				`{} // {type: "@sub"}`,
				[]typ{
					{`@sub`, `{
					id\n: 123
				}`},
				},
				errors.ErrInvalidChildNodeTogetherWithTypeReference,
			},
			{
				`{} // {type: "@sub"}`,
				[]typ{
					{`@sub`, `{
					id": 123
				}`},
				},
				errors.ErrInvalidChildNodeTogetherWithTypeReference,
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
