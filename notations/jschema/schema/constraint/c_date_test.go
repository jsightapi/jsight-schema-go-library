package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestNewDate(t *testing.T) {
	assert.NotNil(t, NewDate())
}

func TestDate_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, NewDate(), json.TypeString)
}

func TestDate_Type(t *testing.T) {
	assert.Equal(t, DateConstraintType, NewDate().Type())
}

func TestDate_String(t *testing.T) {
	assert.Equal(t, DateConstraintType.String(), NewDate().String())
}

func TestDate_Validate(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		NewDate().Validate(bytes.NewBytes("2021-01-08"))
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Date parsing error (parsing time "2021-21-21": month out of range)`, func() {
			NewDate().Validate(bytes.NewBytes("2021-21-21"))
		})
	})
}

func TestDate_ASTNode(t *testing.T) {
	assert.Equal(t, newEmptyRuleASTNode(), NewDate().ASTNode())
}
