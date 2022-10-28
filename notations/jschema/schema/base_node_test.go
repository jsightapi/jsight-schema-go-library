package schema

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestBaseNode_SetRealType(t *testing.T) {
	cc := []struct {
		jsonType    json.Type
		given       string
		expectedRet bool
	}{
		{json.TypeObject, "mixed", true},
		{json.TypeArray, "mixed", true},
		{json.TypeString, "mixed", true},
		{json.TypeInteger, "mixed", true},
		{json.TypeFloat, "mixed", true},
		{json.TypeBoolean, "mixed", true},
		{json.TypeNull, "mixed", true},
		{json.TypeMixed, "mixed", true},

		{json.TypeObject, "enum", false},
		{json.TypeArray, "enum", false},
		{json.TypeString, "enum", true},
		{json.TypeInteger, "enum", true},
		{json.TypeFloat, "enum", true},
		{json.TypeBoolean, "enum", true},
		{json.TypeNull, "enum", true},
		{json.TypeMixed, "enum", false},

		{json.TypeObject, "any", true},
		{json.TypeArray, "any", true},
		{json.TypeString, "any", true},
		{json.TypeInteger, "any", true},
		{json.TypeFloat, "any", true},
		{json.TypeBoolean, "any", true},
		{json.TypeNull, "any", true},
		{json.TypeMixed, "any", true},

		{json.TypeObject, "decimal", false},
		{json.TypeArray, "decimal", false},
		{json.TypeString, "decimal", false},
		{json.TypeInteger, "decimal", false},
		{json.TypeFloat, "decimal", true},
		{json.TypeBoolean, "decimal", false},
		{json.TypeNull, "decimal", false},
		{json.TypeMixed, "decimal", false},

		{json.TypeObject, "email", false},
		{json.TypeArray, "email", false},
		{json.TypeString, "email", true},
		{json.TypeInteger, "email", false},
		{json.TypeFloat, "email", false},
		{json.TypeBoolean, "email", false},
		{json.TypeNull, "email", false},
		{json.TypeMixed, "email", false},

		{json.TypeObject, "uri", false},
		{json.TypeArray, "uri", false},
		{json.TypeString, "uri", true},
		{json.TypeInteger, "uri", false},
		{json.TypeFloat, "uri", false},
		{json.TypeBoolean, "uri", false},
		{json.TypeNull, "uri", false},
		{json.TypeMixed, "uri", false},

		{json.TypeObject, "uuid", false},
		{json.TypeArray, "uuid", false},
		{json.TypeString, "uuid", true},
		{json.TypeInteger, "uuid", false},
		{json.TypeFloat, "uuid", false},
		{json.TypeBoolean, "uuid", false},
		{json.TypeNull, "uuid", false},
		{json.TypeMixed, "uuid", false},

		{json.TypeObject, "date", false},
		{json.TypeArray, "date", false},
		{json.TypeString, "date", true},
		{json.TypeInteger, "date", false},
		{json.TypeFloat, "date", false},
		{json.TypeBoolean, "date", false},
		{json.TypeNull, "date", false},
		{json.TypeMixed, "date", false},

		{json.TypeObject, "datetime", false},
		{json.TypeArray, "datetime", false},
		{json.TypeString, "datetime", true},
		{json.TypeInteger, "datetime", false},
		{json.TypeFloat, "datetime", false},
		{json.TypeBoolean, "datetime", false},
		{json.TypeNull, "datetime", false},
		{json.TypeMixed, "datetime", false},

		{json.TypeObject, "object", true},
		{json.TypeArray, "object", false},
		{json.TypeString, "object", false},
		{json.TypeInteger, "object", false},
		{json.TypeFloat, "object", false},
		{json.TypeBoolean, "object", false},
		{json.TypeNull, "object", false},
		{json.TypeMixed, "object", false},

		{json.TypeObject, "array", false},
		{json.TypeArray, "array", true},
		{json.TypeString, "array", false},
		{json.TypeInteger, "array", false},
		{json.TypeFloat, "array", false},
		{json.TypeBoolean, "array", false},
		{json.TypeNull, "array", false},
		{json.TypeMixed, "array", false},

		{json.TypeObject, "string", false},
		{json.TypeArray, "string", false},
		{json.TypeString, "string", true},
		{json.TypeInteger, "string", false},
		{json.TypeFloat, "string", false},
		{json.TypeBoolean, "string", false},
		{json.TypeNull, "string", false},
		{json.TypeMixed, "string", false},

		{json.TypeObject, "integer", false},
		{json.TypeArray, "integer", false},
		{json.TypeString, "integer", false},
		{json.TypeInteger, "integer", true},
		{json.TypeFloat, "integer", false},
		{json.TypeBoolean, "integer", false},
		{json.TypeNull, "integer", false},
		{json.TypeMixed, "integer", false},

		{json.TypeObject, "float", false},
		{json.TypeArray, "float", false},
		{json.TypeString, "float", false},
		{json.TypeInteger, "float", false},
		{json.TypeFloat, "float", true},
		{json.TypeBoolean, "float", false},
		{json.TypeNull, "float", false},
		{json.TypeMixed, "float", false},

		{json.TypeObject, "boolean", false},
		{json.TypeArray, "boolean", false},
		{json.TypeString, "boolean", false},
		{json.TypeInteger, "boolean", false},
		{json.TypeFloat, "boolean", false},
		{json.TypeBoolean, "boolean", true},
		{json.TypeNull, "boolean", false},
		{json.TypeMixed, "boolean", false},

		{json.TypeObject, "null", false},
		{json.TypeArray, "null", false},
		{json.TypeString, "null", false},
		{json.TypeInteger, "null", false},
		{json.TypeFloat, "null", false},
		{json.TypeBoolean, "null", false},
		{json.TypeNull, "null", true},
		{json.TypeMixed, "null", false},
	}

	for _, c := range cc {
		t.Run(fmt.Sprintf("%s - %s", c.jsonType, c.given), func(t *testing.T) {
			n := &baseNode{
				jsonType: c.jsonType,
			}

			ret := n.SetRealType(c.given)
			if c.expectedRet {
				assert.True(t, ret)
				assert.Equal(t, c.given, n.realType)
			} else {
				assert.False(t, ret)
			}
		})
	}
}

func TestBaseNode_RealType(t *testing.T) {
	n := &baseNode{
		realType: "foo",
	}

	assert.Equal(t, "foo", n.RealType())
}
