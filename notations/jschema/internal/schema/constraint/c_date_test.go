package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestDate_IsJsonTypeCompatible(t *testing.T) {
	cc := map[json.Type]bool{
		json.TypeObject:  false,
		json.TypeArray:   false,
		json.TypeString:  true,
		json.TypeInteger: false,
		json.TypeFloat:   false,
		json.TypeBoolean: false,
		json.TypeNull:    false,
		json.TypeMixed:   false,
	}

	for typ, expected := range cc {
		t.Run(typ.String(), func(t *testing.T) {
			assert.Equal(t, expected, NewDate().IsJsonTypeCompatible(typ))
		})
	}
}

func TestDate_Type(t *testing.T) {
	assert.Equal(t, DateConstraintType, NewDate().Type())
}

func TestDate_String(t *testing.T) {
	assert.Equal(t, DateConstraintType.String(), NewDate().String())
}

func TestDate_Validate(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		NewDate().Validate(bytes.Bytes("2021-01-08"))
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Date parsing error (parsing time "2021-21-21": month out of range)`, func() {
			NewDate().Validate(bytes.Bytes("2021-21-21"))
		})
	})
}

func TestDate_ASTNode(t *testing.T) {
	assert.Equal(t, newEmptyRuleASTNode(), NewDate().ASTNode())
}
