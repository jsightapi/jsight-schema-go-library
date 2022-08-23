package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestNewMax(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		ruleValue := bytes.Bytes("3.14")
		cnstr := NewMax(ruleValue)

		expectedNumber, err := json.NewNumber(ruleValue)
		require.NoError(t, err)

		assert.Equal(t, ruleValue, cnstr.rawValue)
		assert.Equal(t, expectedNumber, cnstr.max)
		assert.False(t, cnstr.exclusive)
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Incorrect number value "not a number"`, func() {
			NewMax([]byte("not a number"))
		})
	})
}

func TestMax_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, Max{}, json.TypeInteger, json.TypeFloat)
}

func TestMax_Type(t *testing.T) {
	assert.Equal(t, MaxConstraintType, NewMax(bytes.Bytes("1")).Type())
}

func TestMax_String(t *testing.T) {
	cc := map[bool]string{
		false: "max: 3.14",
		true:  "max: 3.14 (exclusive: true)",
	}

	for exclusive, expected := range cc {
		t.Run(expected, func(t *testing.T) {
			cnstr := NewMax([]byte("3.14"))
			cnstr.SetExclusive(exclusive)

			actual := cnstr.String()
			assert.Equal(t, expected, actual)
		})
	}
}

func TestMax_SetExclusive(t *testing.T) {
	cnstr := Max{}

	cnstr.SetExclusive(true)
	assert.True(t, cnstr.exclusive)

	cnstr.SetExclusive(false)
	assert.False(t, cnstr.exclusive)
}

func TestMax_Exclusive(t *testing.T) {
	cnstr := Max{}

	cnstr.exclusive = true
	assert.True(t, cnstr.Exclusive())

	cnstr.exclusive = false
	assert.False(t, cnstr.Exclusive())
}

func TestMax_Validate(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		newMax := func(max string, exclusive bool) *Max {
			cnstr := NewMax([]byte(max))
			cnstr.SetExclusive(exclusive)
			return cnstr
		}

		cc := map[string]struct {
			cnstr *Max
			value string
			error string
		}{
			"3.14 <= 3.14": {
				cnstr: newMax("3.14", true),
				value: "3.14",
				error: `Invalid value for "max" = 3.14 constraint (exclusive)`,
			},
			"3.14 <= 2": {
				cnstr: newMax("3.14", true),
				value: "2",
			},
			"3.14 < 3.14": {
				cnstr: newMax("3.14", false),
				value: "3.14",
			},
			"3.14 < 2": {
				cnstr: newMax("3.14", false),
				value: "2",
			},
			"3.14 <= 4": {
				cnstr: newMax("3.14", true),
				value: "4",
				error: `Invalid value for "max" = 3.14 constraint (exclusive)`,
			},
			"3.14 < 4": {
				cnstr: newMax("3.14", false),
				value: "4",
				error: `Invalid value for "max" = 3.14 constraint `,
			},
		}

		for name, c := range cc {
			t.Run(name, func(t *testing.T) {
				if c.error != "" {
					assert.PanicsWithError(t, c.error, func() {
						c.cnstr.Validate([]byte(c.value))
					})
				} else {
					assert.NotPanics(t, func() {
						c.cnstr.Validate([]byte(c.value))
					})
				}
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Incorrect number value "not a number"`, func() {
			NewMax([]byte("3")).Validate([]byte("not a number"))
		})
	})
}

func TestMax_ASTNode(t *testing.T) {
	assert.Equal(t, jschema.RuleASTNode{
		TokenType:  jschema.TokenTypeNumber,
		Value:      "1",
		Properties: &jschema.RuleASTNodes{},
		Source:     jschema.RuleASTNodeSourceManual,
	}, NewMax(bytes.Bytes("1")).ASTNode())
}

func TestMax_Value(t *testing.T) {
	num, err := json.NewNumber([]byte("42"))
	require.NoError(t, err)

	cnstr := Max{
		max: num,
	}
	assert.Equal(t, num, cnstr.Value())
}
