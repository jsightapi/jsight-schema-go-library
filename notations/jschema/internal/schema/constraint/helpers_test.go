package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func testIsJsonTypeCompatible(t *testing.T, cnstr Constraint, compatible ...json.Type) {
	t.Helper()

	cc := map[json.Type]bool{
		json.TypeUndefined: false,
		json.TypeObject:    false,
		json.TypeArray:     false,
		json.TypeString:    false,
		json.TypeInteger:   false,
		json.TypeFloat:     false,
		json.TypeBoolean:   false,
		json.TypeNull:      false,
		json.TypeMixed:     false,
	}

	for _, jsonType := range compatible {
		cc[jsonType] = true
	}

	toString := func(t json.Type) string {
		if t == json.TypeUndefined {
			return "undefined"
		}
		return t.String()
	}

	for jsonType, expected := range cc {
		t.Run(toString(jsonType), func(t *testing.T) {
			actual := cnstr.IsJsonTypeCompatible(jsonType)
			assert.Equal(t, expected, actual)
		})
	}
}

var allJSONTypes = []json.Type{
	json.TypeUndefined,
	json.TypeObject,
	json.TypeArray,
	json.TypeString,
	json.TypeInteger,
	json.TypeFloat,
	json.TypeBoolean,
	json.TypeNull,
	json.TypeMixed,
}
