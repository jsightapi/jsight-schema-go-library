package schema

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidType(t *testing.T) {
	for _, typ := range allSchemaTypes {
		if typ == SchemaTypeUndefined {
			continue
		}

		typ := string(typ)
		t.Run(typ, func(t *testing.T) {
			assert.True(t, IsValidType(typ))
		})
	}

	t.Run("invalid", func(t *testing.T) {
		assert.False(t, IsValidType("invalid"))
	})
}

func TestSchemaType_IsOneOf(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		ttt := [][]SchemaType{
			{SchemaTypeString},
			{SchemaTypeObject, SchemaTypeString, SchemaTypeArray},
		}

		for _, tt := range ttt {
			t.Run(fmt.Sprintf("%v", tt), func(t *testing.T) {
				assert.True(t, SchemaTypeString.IsOneOf(tt...))
			})
		}
	})

	t.Run("false", func(t *testing.T) {
		ttt := [][]SchemaType{
			{},
			{SchemaTypeObject},
			{SchemaTypeObject, SchemaTypeInteger, SchemaTypeArray},
		}

		for _, tt := range ttt {
			t.Run(fmt.Sprintf("%v", tt), func(t *testing.T) {
				assert.False(t, SchemaTypeString.IsOneOf(tt...))
			})
		}

		t.Run("undefined", func(t *testing.T) {
			assert.False(t, SchemaTypeUndefined.IsOneOf(SchemaTypeUndefined))
		})
	})
}

func TestSchemaType_IsEqualSoft(t *testing.T) {
	type testCase struct {
		left     SchemaType
		right    SchemaType
		expected bool
	}

	toString := func(c testCase) string {
		sign := "=="
		if !c.expected {
			sign = "!="
		}
		return fmt.Sprintf("%s %s %s", c.left, sign, c.right)
	}

	notListed := func(ss []SchemaType) []SchemaType {
		if len(ss) == 0 {
			return allSchemaTypes
		}

		ssMap := map[SchemaType]struct{}{}
		for _, s := range ss {
			ssMap[s] = struct{}{}
		}

		res := make([]SchemaType, 0, len(allSchemaTypes)-len(ss))
		for _, s := range allSchemaTypes {
			if _, ok := ssMap[s]; !ok {
				res = append(res, s)
			}
		}
		return res
	}

	var cc []testCase

	for l, rr := range schemaTypeComparisonMap {
		for _, r := range rr {
			cc = append(cc, testCase{l, r, true})
		}
		for _, r := range notListed(rr) {
			cc = append(cc, testCase{l, r, false})
		}
	}

	for _, c := range cc {
		t.Run(toString(c), func(t *testing.T) {
			actual := c.left.IsEqualSoft(c.right)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestGuessSchemaType(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]SchemaType{
			"42":    SchemaTypeInteger,
			"3.14":  SchemaTypeFloat,
			`"foo"`: SchemaTypeString,
			"true":  SchemaTypeBoolean,
			"false": SchemaTypeBoolean,
			"{":     SchemaTypeObject,
			"[":     SchemaTypeArray,
			"null":  SchemaTypeNull,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				actual, err := GuessSchemaType([]byte(given))
				require.NoError(t, err)
				assert.Equal(t, expected, actual)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		_, err := GuessSchemaType([]byte("invalid"))
		assert.ErrorIs(t, err, ErrUnknownSchemaType)
	})
}

var allSchemaTypes = []SchemaType{
	SchemaTypeUndefined,
	SchemaTypeString,
	SchemaTypeInteger,
	SchemaTypeFloat,
	SchemaTypeDecimal,
	SchemaTypeBoolean,
	SchemaTypeObject,
	SchemaTypeArray,
	SchemaTypeNull,
	SchemaTypeEmail,
	SchemaTypeURI,
	SchemaTypeUUID,
	SchemaTypeDate,
	SchemaTypeDateTime,
	SchemaTypeEnum,
	SchemaTypeMixed,
	SchemaTypeAny,
	SchemaTypeComment,
}
