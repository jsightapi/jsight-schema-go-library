package constraint

import (
	"github.com/stretchr/testify/assert"
	"j/schema/bytes"
	"testing"
)

func TestUuid_Type(t *testing.T) {
	assert.Equal(t, UuidConstraintType, NewUuid().Type())
}

func TestUuid_Validate(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		var tests = []string{
			`550e8400-e29b-41d4-a716-446655440000`,
			`urn:uuid:550e8400-e29b-41d4-a716-446655440000`,
			`URN:UUID:550e8400-e29b-41d4-a716-446655440000`,
			`{550e8400-e29b-41d4-a716-446655440000}`,
			`550e8400e29b41d4a716446655440000`,
			`aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee`,
			`AAAAAAAA-BBBB-CCCC-DDDD-EEEEEEEEEEEE`,
		}

		for _, value := range tests {
			t.Run(value, func(t *testing.T) {
				NewUuid().Validate(bytes.Bytes(value))
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		var tests = []string{
			``,
			`12`,
			`1.2`,
			`true`,
			`false`,
			`null`,
			`"ABC"`,
			` 550e8400e29b41d4a716446655440000`,      // leading symbol " "
			` 550e8400-e29b-41d4-a716-446655440000`,  // leading symbol " "
			`550e8400e29b41d4a716446655440000 `,      // trailing symbol " "
			`550e8400-e29b-41d4-a716-446655440000 `,  // trailing symbol " "
			` 550e8400e29b41d4a716446655440000 `,     // leading  and trailing symbol " "
			` 550e8400-e29b-41d4-a716-446655440000 `, // leading  and trailing symbol " "
			`550e8400e29b41d4a7164466554400009`,      // additional trailing symbol "9"
			`550e840-0e29b-41d4-a716-446655440000`,   // invalid symbol "-" location
			`z50e8400-e29b-41d4-a716-446655440000`,   // invalid symbol "z"
		}

		for _, value := range tests {
			t.Run(value, func(*testing.T) {
				assert.Panics(t, func() {
					NewUuid().Validate(bytes.Bytes(value))
				})
			})
		}
	})
}

func TestUuid_ASTNode(t *testing.T) {
	assert.Equal(t, newEmptyRuleASTNode(), Uuid{}.ASTNode())
}
