package constraint

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	jbytes "github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
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
				mode:     AdditionalPropertiesMustBeType,
				typeName: jbytes.Bytes("@type"),
			},

			"object": {
				mode:       AdditionalPropertiesMustBeSchemaType,
				schemaType: jschema.SchemaTypeObject,
			},
		}

		for v, expected := range cc {
			t.Run(v, func(t *testing.T) {
				t.Parallel()

				actual := NewAdditionalProperties(jbytes.Bytes(v))
				assert.True(t, actual.IsEqual(expected))
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Unknown JSchema type "foo"`, func() {
			NewAdditionalProperties(jbytes.Bytes("foo"))
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
				schemaType: jschema.SchemaTypeObject,
			},
			"additionalProperties: @foo": {
				mode:     AdditionalPropertiesMustBeType,
				typeName: jbytes.Bytes("@foo"),
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
		AdditionalPropertiesMustBeType,
	}

	for _, m := range cc {
		t.Run(strconv.Itoa(int(m)), func(t *testing.T) {
			actual := AdditionalProperties{mode: m}.Mode()
			assert.Equal(t, m, actual)
		})
	}
}

func TestAdditionalProperties_JsonType(t *testing.T) {
	const expected = jschema.SchemaTypeArray

	actual := AdditionalProperties{schemaType: expected}.SchemaType()
	assert.Equal(t, actual, expected)
}

func TestAdditionalProperties_TypeName(t *testing.T) {
	var expected = jbytes.Bytes("@foo")

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
				schemaType: jschema.SchemaTypeObject,
				typeName:   jbytes.Bytes("foo"),
			},
			AdditionalProperties{
				schemaType: jschema.SchemaTypeObject,
				typeName:   jbytes.Bytes("foo"),
			},
			true,
		},

		"same but with different modes": {
			AdditionalProperties{
				mode:       AdditionalPropertiesMustBeType,
				schemaType: jschema.SchemaTypeObject,
				typeName:   jbytes.Bytes("foo"),
			},
			AdditionalProperties{
				mode:       AdditionalPropertiesMustBeSchemaType,
				schemaType: jschema.SchemaTypeObject,
				typeName:   jbytes.Bytes("foo"),
			},
			true,
		},

		"different": {
			AdditionalProperties{
				typeName: jbytes.Bytes("foo"),
			},
			AdditionalProperties{
				schemaType: jschema.SchemaTypeObject,
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
	cc := map[string]jschema.RuleASTNode{
		`"any"`: {
			TokenType:  jschema.TokenTypeString,
			Value:      "any",
			Properties: &jschema.RuleASTNodes{},
			Source:     jschema.RuleASTNodeSourceManual,
		},

		"true": {
			TokenType:  jschema.TokenTypeBoolean,
			Value:      "true",
			Properties: &jschema.RuleASTNodes{},
			Source:     jschema.RuleASTNodeSourceManual,
		},

		"false": {
			TokenType:  jschema.TokenTypeBoolean,
			Value:      "false",
			Properties: &jschema.RuleASTNodes{},
			Source:     jschema.RuleASTNodeSourceManual,
		},

		`"@foo"`: {
			TokenType:  jschema.TokenTypeString,
			Value:      "@foo",
			Properties: &jschema.RuleASTNodes{},
			Source:     jschema.RuleASTNodeSourceManual,
		},

		`"string"`: {
			TokenType:  jschema.TokenTypeString,
			Value:      "string",
			Properties: &jschema.RuleASTNodes{},
			Source:     jschema.RuleASTNodeSourceManual,
		},

		`"integer"`: {
			TokenType:  jschema.TokenTypeString,
			Value:      "integer",
			Properties: &jschema.RuleASTNodes{},
			Source:     jschema.RuleASTNodeSourceManual,
		},
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			assert.Equal(t, expected, NewAdditionalProperties([]byte(given)).ASTNode())
		})
	}
}
