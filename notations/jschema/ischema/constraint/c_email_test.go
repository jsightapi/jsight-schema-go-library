package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/json"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestEmail_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, NewEmail(), json.TypeString)
}

func TestEmail_Type(t *testing.T) {
	assert.Equal(t, EmailConstraintType, NewEmail().Type())
}

func TestEmail_String(t *testing.T) {
	assert.Equal(t, "email", NewEmail().String())
}

func TestEmail_Validate(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		var tests = []string{
			`prettyandsimple@example.com`,
			`very.common@example.com`,
			`disposable.style.email.with+symbol@example.com`,
			`other.email-with-dash@example.com`,
			`x@example.com`,
			`"John..Doe"@example.com`,
			`"much.more unusual"@example.com`,
			`"very.unusual.@.unusual.com"@example.com`,
			`"very.(),:;<>[]\".VERY.\"very@\ \"very\".unusual"@strange.example.com`,
			`example-indeed@strange-example.com`,
			`admin@mailserver1`,
			"#!$%&'*+-/=?^_`{}|~@example.org",
			`" "@example.org`,
			`example@localhost`,
			`example@s.solutions`,
			`user@com`,
			`user@localserver`,
			`ç$€§/az@gmail.com`,
		}

		for _, email := range tests {
			t.Run(email, func(t *testing.T) {
				NewEmail().Validate(bytes.NewBytes(email))
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		var tests = map[string]string{
			"":   "Empty email",
			`""`: "Empty email",
			// no @ character
			"Abc.example.com": "Invalid email (Abc.example.com)",
			// only one @ is allowed outside quotation marks
			"A@b@c@example.com": "Invalid email (A@b@c@example.com)",
			// none of the special characters in this local part are allowed outside quotation marks
			`a"b(c)d,e:f;gi[j\k]l@example.com`: `Invalid email (a"b(c)d,e:f;gi[j\k]l@example.com)`,
			// quoted strings must be dot separated or the only element making up the local part
			`just"not"right@example.com`: `Invalid email (just"not"right@example.com)`,
			// spaces, quotes, and backslashes may only exist when within quoted strings and preceded by a backslash
			`this is"not\allowed@example.com`: `Invalid email (this is"not\allowed@example.com)`,
			// even if escaped (preceded by a backslash), spaces, quotes, and backslashes must still be contained by quotes
			`this\ still\"not\allowed@example.com`: `Invalid email (this\ still\"not\allowed@example.com)`,
			// double dot before @; (with caveat: Gmail lets this through)
			"john..doe@example.com": `Invalid email (john..doe@example.com)`,
			// double dot after @
			"john.doe@example..com": `Invalid email (john.doe@example..com)`,

			// a valid address with name
			"Barry Gibbs <bg@example.com>": "Invalid email (Barry Gibbs <bg@example.com>)",
			// a valid address with a leading space
			" aaa@bbb.cc": "Invalid email ( aaa@bbb.cc)",
			// a valid address with a trailing space
			"aaa@bbb.cc ":      "Invalid email (aaa@bbb.cc )",
			"bg@example.com>":  "Invalid email (bg@example.com>)",
			"<bg@example.com":  "Invalid email (<bg@example.com)",
			"<bg@example.com>": "Invalid email (<bg@example.com>)",
		}

		for email, expected := range tests {
			t.Run(email, func(t *testing.T) {
				assert.PanicsWithError(t, expected, func() {
					NewEmail().Validate(bytes.NewBytes(email))
				})
			})
		}
	})
}

func TestEmail_ASTNode(t *testing.T) {
	assert.Equal(t, newEmptyRuleASTNode(), Email{}.ASTNode())
}
