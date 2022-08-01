package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/internal/lexeme"
)

func TestNewConstraintFromRule(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]struct {
			val          string
			expectedType Constraint
		}{
			"minLength":            {"1", &MinLength{}},
			"maxLength":            {"1", &MaxLength{}},
			"min":                  {"1", &Min{}},
			"max":                  {"1", &Max{}},
			"exclusiveMinimum":     {"true", &ExclusiveMinimum{}},
			"exclusiveMaximum":     {"true", &ExclusiveMaximum{}},
			"type":                 {"foo", &TypeConstraint{}},
			"precision":            {"1", &Precision{}},
			"optional":             {"true", &Optional{}},
			"minItems":             {"1", &MinItems{}},
			"maxItems":             {"1", &MaxItems{}},
			"additionalProperties": {"true", &AdditionalProperties{}},
			"nullable":             {"true", &Nullable{}},
			"regex":                {`"."`, &Regex{}},
			"const":                {"true", &Const{}},
		}

		for given, c := range cc {
			t.Run(given, func(t *testing.T) {
				constraint := NewConstraintFromRule(
					lexeme.NewLexEvent(
						lexeme.LiteralBegin,
						0,
						bytes.Index(len(given))-1,
						fs.NewFile("", given),
					),
					bytes.Bytes(c.val),
					nil,
				)

				assert.IsType(t, c.expectedType, constraint)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `ERROR (code 601): Unknown rule "invalid"
	in line 1 on file 
	> invalid
	--^`, func() {
			const given = "invalid"

			NewConstraintFromRule(
				lexeme.NewLexEvent(
					lexeme.LiteralBegin,
					0,
					bytes.Index(len(given))-1,
					fs.NewFile("", given),
				),
				nil,
				nil,
			)
		})
	})
}
