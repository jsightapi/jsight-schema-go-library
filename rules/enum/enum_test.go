package enum

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestEnum_Len(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]uint{
			"[]":               2,
			"[]   \t  \n  \r ": 2,
			`[
	42,
	3.14,
	"foo",
	true,
	false,
	null
]`: 44,
			"[42] something": 4,
			"":               0,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				actual, err := New("", []byte(given)).Len()
				require.NoError(t, err)
				assert.Equal(t, expected, actual)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]string{
			`ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file 
	> 42
	--^`: "42",

			`ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file 
	> 42 [] foo
	--^`: "42 [] foo",

			`ERROR (code 303): Unexpected end of file
	in line 1 on file 
	> [
	--^`: "[",
		}

		for expected, given := range cc {
			t.Run(expected, func(t *testing.T) {
				_, err := New("", []byte(given)).Len()
				assert.EqualError(t, err, expected)
			})
		}
	})
}

func TestEnum_Check(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		testList := []string{
			"[]",
			"[1]",
			"[1,2]",
			"[1,2,3]",
			"   [1,2,3]   ",
			"   [1,  2,  3]   ",
			"\n[1,2]",
			"[\n1,2]",
			"[1\n,2]",
			"[1,\n2]",
			"[1,2\n]",
			"[1,2]\n",
			`["aaa", "bbb", "ccc"]`,
			`[123, 45.67, "abc", true, false, null]`,
			`[
	123,
	45.67,
	"abc",
	true,
	false,
	null
]`,
		}

		for _, enum := range testList {
			t.Run(enum, func(t *testing.T) {
				err := New("enum", []byte(enum)).Check()
				require.NoError(t, err)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]string{
			"123": `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file enum
	> 123
	--^`,
			`"abc"`: `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file enum
	> "abc"
	--^`,
			"true": `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file enum
	> true
	--^`,
			"false": `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file enum
	> false
	--^`,
			"null": `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file enum
	> null
	--^`,
			"{}": `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file enum
	> {}
	--^`,
			"[1,2,3] xxx": `ERROR (code 301): Invalid character "x" non-space byte after top-level value
	in line 1 on file enum
	> [1,2,3] xxx
	----------^`,
			"xxx [1,2,3]": `ERROR (code 1600): An array was expected as a value for the "enum"
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
			"[ {} ]": `ERROR (code 301): Invalid character "{" looking for beginning of value
	in line 1 on file enum
	> [ {} ]
	----^`,
			"[ [] ]": `ERROR (code 301): Invalid character "[" looking for beginning of value
	in line 1 on file enum
	> [ [] ]
	----^`,
		}

		for enum, expected := range cc {
			t.Run(enum, func(t *testing.T) {
				err := New("enum", []byte(enum)).Check()
				assert.EqualError(t, err, expected)
			})
		}
	})
}

func TestEnum_Values(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		vals, err := New("", []byte(`[
	"foo",
	42,
	3.14,
	true,
	false,
	null
]`)).
			Values()

		require.NoError(t, err)
		assert.Equal(t, []bytes.Bytes{
			[]byte(`"foo"`),
			[]byte("42"),
			[]byte("3.14"),
			[]byte("true"),
			[]byte("false"),
			[]byte("null"),
		}, vals)
	})

	t.Run("negative", func(t *testing.T) {
		_, err := New("", []byte("123")).Values()
		assert.EqualError(t, err, `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file 
	> 123
	--^`)
	})
}
