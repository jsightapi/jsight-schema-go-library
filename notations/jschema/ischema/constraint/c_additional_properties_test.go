package constraint

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/json"

	schema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestNewAdditionalProperties(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]AdditionalProperties{
			"any": {
				mode: AdditionalPropertiesCanBeAny,
			},

			"true": {
				mode: AdditionalPropertiesCanBeAny,
			},

			"false": {
				mode: AdditionalPropertiesNotAllowed,
			},

			"@type": {
				mode:     AdditionalPropertiesMustBeUserType,
				typeName: bytes.NewBytes("@type"),
			},

			"object": {
				mode:       AdditionalPropertiesMustBeSchemaType,
				schemaType: schema.SchemaTypeObject,
			},
		}

		for v, expected := range cc {
			t.Run(v, func(t *testing.T) {
				t.Parallel()

				actual := NewAdditionalProperties(bytes.NewBytes(v))
				assert.True(t, actual.IsEqual(expected))
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Unknown JSchema type "foo"`, func() {
			NewAdditionalProperties(bytes.NewBytes("foo"))
		})
	})
}

func TestAdditionalProperties_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, AdditionalProperties{}, json.TypeObject)
}

func TestAdditionalProperties_Type(t *testing.T) {
	const expected = AdditionalPropertiesConstraintType

	actual := AdditionalProperties{}.Type()
	assert.Equal(t, expected, actual)
}

func TestAdditionalProperties_String(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]AdditionalProperties{
			"additionalProperties: any": {
				mode: AdditionalPropertiesCanBeAny,
			},
			"additionalProperties: object": {
				mode:       AdditionalPropertiesMustBeSchemaType,
				schemaType: schema.SchemaTypeObject,
			},
			"additionalProperties: @foo": {
				mode:     AdditionalPropertiesMustBeUserType,
				typeName: bytes.NewBytes("@foo"),
			},
			"additionalProperties: false": {
				mode: AdditionalPropertiesNotAllowed,
			},
		}

		for expected, p := range cc {
			t.Run(expected, func(t *testing.T) {
				actual := p.String()
				assert.Equal(t, expected, actual)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, "Constraint error", func() {
			_ = AdditionalProperties{
				mode: -1,
			}.String()
		})
	})
}

func TestAdditionalProperties_Mode(t *testing.T) {
	cc := []AdditionalPropertiesMode{
		AdditionalPropertiesCanBeAny,
		AdditionalPropertiesMustBeSchemaType,
		AdditionalPropertiesMustBeUserType,
	}

	for _, m := range cc {
		t.Run(strconv.Itoa(int(m)), func(t *testing.T) {
			actual := AdditionalProperties{mode: m}.Mode()
			assert.Equal(t, m, actual)
		})
	}
}

func TestAdditionalProperties_JsonType(t *testing.T) {
	const expected = schema.SchemaTypeArray

	actual := AdditionalProperties{schemaType: expected}.SchemaType()
	assert.Equal(t, actual, expected)
}

func TestAdditionalProperties_TypeName(t *testing.T) {
	var expected = bytes.NewBytes("@foo")

	actual := AdditionalProperties{typeName: expected}.TypeName()
	assert.Equal(t, expected, actual)
}

func TestAdditionalProperties_IsEqual(t *testing.T) {
	cc := map[string]struct {
		c1, c2   AdditionalProperties
		expected bool
	}{
		"two empty": {AdditionalProperties{}, AdditionalProperties{}, true},

		"same": {
			AdditionalProperties{
				schemaType: schema.SchemaTypeObject,
				typeName:   bytes.NewBytes("foo"),
			},
			AdditionalProperties{
				schemaType: schema.SchemaTypeObject,
				typeName:   bytes.NewBytes("foo"),
			},
			true,
		},

		"same but with different modes": {
			AdditionalProperties{
				mode:       AdditionalPropertiesMustBeUserType,
				schemaType: schema.SchemaTypeObject,
				typeName:   bytes.NewBytes("foo"),
			},
			AdditionalProperties{
				mode:       AdditionalPropertiesMustBeSchemaType,
				schemaType: schema.SchemaTypeObject,
				typeName:   bytes.NewBytes("foo"),
			},
			true,
		},

		"different": {
			AdditionalProperties{
				typeName: bytes.NewBytes("foo"),
			},
			AdditionalProperties{
				schemaType: schema.SchemaTypeObject,
			},
			false,
		},
	}

	for n, c := range cc {
		t.Run(n, func(t *testing.T) {
			t.Parallel()

			actual := c.c1.IsEqual(c.c2)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestAdditionalProperties_ASTNode(t *testing.T) {
	cc := map[string]schema.RuleASTNode{
		`"any"`: {
			TokenType:  schema.TokenTypeString,
			Value:      "any",
			Properties: &schema.RuleASTNodes{},
			Source:     schema.RuleASTNodeSourceManual,
		},

		"true": {
			TokenType:  schema.TokenTypeBoolean,
			Value:      "true",
			Properties: &schema.RuleASTNodes{},
			Source:     schema.RuleASTNodeSourceManual,
		},

		"false": {
			TokenType:  schema.TokenTypeBoolean,
			Value:      "false",
			Properties: &schema.RuleASTNodes{},
			Source:     schema.RuleASTNodeSourceManual,
		},

		`"@foo"`: {
			TokenType:  schema.TokenTypeString,
			Value:      "@foo",
			Properties: &schema.RuleASTNodes{},
			Source:     schema.RuleASTNodeSourceManual,
		},

		`"string"`: {
			TokenType:  schema.TokenTypeString,
			Value:      "string",
			Properties: &schema.RuleASTNodes{},
			Source:     schema.RuleASTNodeSourceManual,
		},

		`"integer"`: {
			TokenType:  schema.TokenTypeString,
			Value:      "integer",
			Properties: &schema.RuleASTNodes{},
			Source:     schema.RuleASTNodeSourceManual,
		},
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			assert.Equal(t, expected, NewAdditionalProperties(bytes.NewBytes(given)).ASTNode())
		})
	}
}
