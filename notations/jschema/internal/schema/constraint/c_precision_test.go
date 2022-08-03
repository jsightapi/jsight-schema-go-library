package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestNewPrecision(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		c := NewPrecision([]byte("10"))
		assert.Equal(t, uint(10), c.value)
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]string{
			"-10":  `Invalid value of "precision" constraint`,
			"0":    "Precision can't be zero",
			"3.14": `Invalid value of "precision" constraint`,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				assert.PanicsWithError(t, expected, func() {
					NewPrecision([]byte(given))
				})
			})
		}
	})
}

func TestPrecision_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, Precision{}, json.TypeFloat)
}

func TestPrecision_Type(t *testing.T) {
	assert.Equal(t, PrecisionConstraintType, NewPrecision(bytes.Bytes("1")).Type())
}

func TestPrecision_String(t *testing.T) {
	assert.Equal(t, "precision: 1", NewPrecision([]byte("1")).String())
}

func TestPrecision_Validate(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]string{
			"3.14":  "",
			"3.1":   "",
			"3":     "",
			"3.142": `Invalid value for "precision" = 2 constraint (exclusive)`,
		}

		for given, expectedError := range cc {
			t.Run(given, func(t *testing.T) {
				cnstr := NewPrecision([]byte("2"))
				if expectedError != "" {
					assert.PanicsWithError(t, expectedError, func() {
						cnstr.Validate([]byte(given))
					})
				} else {
					assert.NotPanics(t, func() {
						cnstr.Validate([]byte(given))
					})
				}
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Incorrect number value "not a number"`, func() {
			NewPrecision([]byte("2")).Validate([]byte("not a number"))
		})
	})
}

func TestPrecision_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		JSONType:   jschema.JSONTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewPrecision(bytes.Bytes("1")).ASTNode())
}
