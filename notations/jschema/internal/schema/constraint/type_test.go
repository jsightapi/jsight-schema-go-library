package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestType_String(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[Type]string{
			MinLengthConstraintType:            "minLength",
			MaxLengthConstraintType:            "maxLength",
			MinConstraintType:                  "min",
			MaxConstraintType:                  "max",
			ExclusiveMinimumConstraintType:     "exclusiveMinimum",
			ExclusiveMaximumConstraintType:     "exclusiveMaximum",
			PrecisionConstraintType:            "precision",
			TypeConstraintType:                 "type",
			TypesListConstraintType:            "types",
			OptionalConstraintType:             "optional",
			OrConstraintType:                   "or",
			RequiredKeysConstraintType:         "required-keys",
			EmailConstraintType:                "email",
			MinItemsConstraintType:             "minItems",
			MaxItemsConstraintType:             "maxItems",
			EnumConstraintType:                 "enum",
			AdditionalPropertiesConstraintType: "additionalProperties",
			AllOfConstraintType:                "allOf",
			AnyConstraintType:                  "any",
			NullableConstraintType:             "nullable",
			RegexConstraintType:                "regex",
			UriConstraintType:                  "uri",
			DateConstraintType:                 "date",
			DateTimeConstraintType:             "datetime",
			UuidConstraintType:                 "uuid",
			ConstConstraintType:                "const",
		}

		for typ, expected := range cc {
			t.Run(expected, func(t *testing.T) {
				assert.Equal(t, expected, typ.String())
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithValue(t, "Unknown constraint type", func() {
			_ = Type(-1).String()
		})
	})
}
