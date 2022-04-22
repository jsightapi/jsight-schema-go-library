package jschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckEnum(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		testList := []string{
			`[]`,
			`[1]`,
			`[1,2]`,
			`[1,2,3]`,
			`   [1,2,3]   `,
			`   [1,  2,  3]   `,
			"\n[1,2]",
			"[\n1,2]",
			"[1\n,2]",
			"[1,\n2]",
			"[1,2\n]",
			"[1,2]\n",
			`["aaa", "bbb", "ccc"]`,
			`[123, 45.67, "abc", true, false, null] `,
		}

		for _, enum := range testList {
			t.Run(enum, func(t *testing.T) {
				err := NewEnum("enum", []byte(enum)).Check()
				require.NoError(t, err)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]string{
			"123": `ERROR (code 806): An array was expected as a value for the "enum"
	in line 1 on file enum
	> 123
	--^`,
			`"abc"`: `ERROR (code 806): An array was expected as a value for the "enum"
	in line 1 on file enum
	> "abc"
	--^`,
			"true": `ERROR (code 806): An array was expected as a value for the "enum"
	in line 1 on file enum
	> true
	--^`,
			"false": `ERROR (code 806): An array was expected as a value for the "enum"
	in line 1 on file enum
	> false
	--^`,
			"null": `ERROR (code 806): An array was expected as a value for the "enum"
	in line 1 on file enum
	> null
	--^`,
			"{}": `ERROR (code 806): An array was expected as a value for the "enum"
	in line 1 on file enum
	> {}
	--^`,
			"[1,2,3] xxx": `ERROR (code 301): Invalid character "x" non-space byte after top-level value
	in line 1 on file enum
	> [1,2,3] xxx
	----------^`,
			"xxx [1,2,3]": `ERROR (code 301): Invalid character "x" looking for beginning of value
	in line 1 on file enum
	> xxx [1,2,3]
	--^`,
			"[1,]": `ERROR (code 301): Invalid character "]" looking for beginning of value
	in line 1 on file enum
	> [1,]
	-----^`,
			"[,1]": `ERROR (code 301): Invalid character "," looking for beginning of value
	in line 1 on file enum
	> [,1]
	---^`,
			"[ {} ]": `ERROR (code 807): Incorrect array item type in "enum". Only literals are allowed.
	in line 1 on file enum
	> [ {} ]
	----^`,
			"[ [] ]": `ERROR (code 807): Incorrect array item type in "enum". Only literals are allowed.
	in line 1 on file enum
	> [ [] ]
	----^`,
		}

		for enum, expected := range cc {
			t.Run(enum, func(t *testing.T) {
				err := NewEnum("enum", []byte(enum)).Check()
				assert.EqualError(t, err, expected)
			})
		}
	})
}
