package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestNewDateTime(t *testing.T) {
	assert.NotNil(t, NewDateTime())
}

func TestDateTime_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, NewDateTime(), json.TypeString)
}

func TestDateTime_Type(t *testing.T) {
	assert.Equal(t, DateTimeConstraintType, NewDateTime().Type())
}

func TestDateTime_String(t *testing.T) {
	assert.Equal(t, DateTimeConstraintType.String(), NewDateTime().String())
}

func TestDateTime_Validate(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := []string{
			"2006-01-02T15:04:05+07:00",
			"2011-10-08T23:11:44-01:00",
			"2011-10-08T23:11:44Z",
		}

		for _, c := range cc {
			t.Run(c, func(t *testing.T) {
				assert.NotPanics(t, func() {
					NewDateTime().Validate(bytes.NewBytes(c))
				})
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		var tests = []string{
			"",
			"12",
			"1.2",
			"true",
			"false",
			"null",
			`"ABC"`,
			" 02 Jan 2006 15:04:05 -0700", // space before date
			"02 Jan 2006 15:04:05 -0700 ", // space after date
			"32 Jan 2006 15:04:05 -07000", // an extra zero after the zone value
			"32 Jan 2006 15:04:05 -0700 ", // day out of range
			"29 Feb 2019 23:59:05 -0300",  // day out of range
			"02 Jan -2006 15:04:05 -0700", // the negative value of the year
			"2 Jan 2006 15:04:05 -0700",   // the leading zero in the date is missing
			"02 Jan 06 15:04:05 -0700",    // year is absent
			"02 Jan 2006 15:04:05 -07:00", // colon
			"02 Jan 2006 15:4:05 -0700",   // cannot parse "4" as "04"
			"02 Jan 2006 3:04:5 -0700",    // cannot parse "5" as "05"
			"02 Jan 2006 15:04:05",        // zone is absent
			"02 Jan 2006",                 // year and zone is absent
			"02 Jan 2006 15:04:05 -07",    // no trailing zeros
			"02 Jan 2006 24:00:00 -0700",  // It's not possible to specify midnight as 24:00
			"2011-10-08T23:11:44",
		}

		for _, value := range tests {
			t.Run(value, func(t *testing.T) {
				assert.PanicsWithValue(t, errors.ErrInvalidDateTime, func() {
					NewDateTime().Validate(bytes.NewBytes(value))
				})
			})
		}
	})
}

func TestDateTime_ASTNode(t *testing.T) {
	assert.Equal(t, newEmptyRuleASTNode(), NewDateTime().ASTNode())
}
