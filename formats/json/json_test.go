package json

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jsightapi/jsight-schema-go-library/reader"
	"github.com/jsightapi/jsight-schema-go-library/test"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkDocument_NextLexeme(b *testing.B) {
	file := reader.Read(filepath.Join(test.GetProjectRoot(), "testdata", "big.json"))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := FromFile(file, AllowTrailingNonSpaceCharacters())
		for {
			_, err := s.NextLexeme()
			if errors.Is(err, io.EOF) {
				break
			}
			require.NoError(b, err)
		}
	}
}

func TestDocument_Len(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				actual, err := MustNew("", c.data, AllowTrailingNonSpaceCharacters()).Len()
				require.NoError(t, err)
				assert.Equal(t, c.expectedLen, actual)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		_, err := MustNew("", "foo", AllowTrailingNonSpaceCharacters()).Len()
		assert.EqualError(t, err, `ERROR (code 301): Invalid character "o" in literal false (expecting 'a')
	in line 1 on file 
	> foo
	---^`)
	})
}

func BenchmarkDocument_Len(b *testing.B) {
	file := reader.Read(filepath.Join(test.GetProjectRoot(), "testdata", "big.json"))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := FromFile(file, AllowTrailingNonSpaceCharacters())
		_, err := s.Len()
		require.NoError(b, err)
	}
}

func TestDocument_Check(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		for name, c := range cases {
			t.Run(name, func(t *testing.T) {
				err := MustNew("", c.data, AllowTrailingNonSpaceCharacters()).Check()
				require.NoError(t, err)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		t.Run("invalid character", func(t *testing.T) {
			err := MustNew("", "foo", AllowTrailingNonSpaceCharacters()).Check()
			assert.EqualError(t, err, `ERROR (code 301): Invalid character "o" in literal false (expecting 'a')
	in line 1 on file 
	> foo
	---^`)
		})

		t.Run("without allowed trailing non empty character", func(t *testing.T) {
			cc := []string{
				`42
some trailing data`,
				`-42
some trailing data`,
				`3.14
some trailing data`,
				`-3.14
some trailing data`,
				`314e-2
some trailing data`,
				`0.314e+1
some trailing data`,
				`3.14e0
some trailing data`,
				`314E-2
some trailing data`,
				`0.314E+1
some trailing data`,
				`3.14E0
some trailing data`,
				`0.14
some trailing data`,
				`-0.14
some trailing data`,
				`true
some trailing data`,
				`false
some trailing data`,
				`null
some trailing data`,
				`"str"
some trailing data`,
				`[
1,
2,
3
]
some trailing data`,
				`
{
	"foo": "bar"
}
some trailing data`,
			}

			for _, given := range cc {
				t.Run(given, func(t *testing.T) {
					err := MustNew("", given).Check()
					assert.True(t, strings.HasPrefix(
						err.Error(),
						"ERROR (code 301): Invalid character \"s\" non-space byte after top-level value",
					))
				})
			}
		})
	})
}

func BenchmarkDocument_Check(b *testing.B) {
	file := reader.Read(filepath.Join(test.GetProjectRoot(), "testdata", "big.json"))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := FromFile(file, AllowTrailingNonSpaceCharacters()).Check()
		require.NoError(b, err)
	}
}

var cases = map[string]struct {
	data        string
	expectedLen uint
}{
	"integer": {"42", 2},
	"integer_with_trailing_data": {`42
some trailing data`, 2},
	"negative_integer": {"-42", 3},
	"negative_integer_with_trailing_data": {`-42
some trailing data`, 3},
	"float": {"3.14", 4},
	"float_with_trailing_data": {`3.14
some trailing data`, 4},
	"negative_float": {"-3.14", 5},
	"negative_float_with_trailing_data": {`-3.14
some trailing data`, 5},
	"exponent_1": {"314e-2", 6},
	"exponent_1_with_trailing_data": {`314e-2
some trailing data`, 6},
	"exponent_2": {"0.314e+1", 8},
	"exponent_2_with_trailing_data": {`0.314e+1
some trailing data`, 8},
	"exponent_3": {"3.14e0", 6},
	"exponent_3_with_trailing_data": {`3.14e0
some trailing data`, 6},
	"exponent_4": {"314E-2", 6},
	"exponent_4_with_trailing_data": {`314E-2
some trailing data`, 6},
	"exponent_5": {"0.314E+1", 8},
	"exponent_5_with_trailing_data": {`0.314E+1
some trailing data`, 8},
	"exponent_6": {"3.14E0", 6},
	"exponent_6_with_trailing_data": {`3.14E0
some trailing data`, 6},
	"zero_beginning_float": {"0.14", 4},
	"zero_beginning_float_with_trailing_data": {`0.14
some trailing data`, 4},
	"negative_zero_beginning_float": {"-0.14", 5},
	"negative_zero_beginning_float_with_trailing_data": {`-0.14
some trailing data`, 5},
	"boolean_true": {"true", 4},
	"boolean_true_with_trailing_data": {`true
some trailing data`, 4},
	"boolean_false": {"false", 5},
	"boolean_false_with_trailing_data": {`false
some trailing data`, 5},
	"nullable": {"null", 4},
	"nullable_with_trailing_data": {`null
some trailing data`, 4},
	"string": {`"str"`, 5},
	"string_with_trailing_data": {`"str"
some trailing data`, 5},

	"array": {`
[
	42,
	-42,
	3.14,
	-3.14,
	314e-2,
	0.314e+1,
	3.14e0,
	314E-2,
	0.314E+1,
	3.14E0,
	0.14,
	-0.14,
	true,
	false,
	null,
	"str",
	{
		"integer": 42,
		"negative_integer": -42,
		"float": 3.14,
		"negative_float": -3.14,
		"exponent_1": 314e-2,
		"exponent_2": 0.314e+1,
		"exponent_3": 3.14e0,
		"exponent_4": 314E-2,
		"exponent_5": 0.314E+1,
		"exponent_6": 3.14E0,
		"zero_beginning_float": 0.14,
		"negative_zero_beginning_float": -0.14,
		"boolean_true": true,
		"boolean_false": false,
		"nullable": null,
		"string": "str",
		"array": [
			42,
			-42,
			3.14,
			-3.14,
			314e-2,
			0.314e+1,
			3.14e0,
			314E-2,
			0.314E+1,
			3.14E0,
			0.14,
			-0.14,
			true,
			false,
			null,
			"str",
			{"object": {}}
		]
	}
]`, 734},
	"array_with_trailing_data": {`[
1,
2,
3
]
some trailing data`, 11},
	"object": {`
{
	"integer": 42,
	"negative_integer": -42,
	"float": 3.14,
	"negative_float": -3.14,
	"exponent_1": 314e-2,
	"exponent_2": 0.314e+1,
	"exponent_3": 3.14e0,
	"exponent_4": 314E-2,
	"exponent_5": 0.314E+1,
	"exponent_6": 3.14E0,
	"zero_beginning_float": 0.14,
	"negative_zero_beginning_float": -0.14,
	"boolean_true": true,
	"boolean_false": false,
	"nullable": null,
	"string": "str",
	"array": [
		42,
		-42,
		3.14,
		-3.14,
		314e-2,
		0.314e+1,
		3.14e0,
		314E-2,
		0.314E+1,
		3.14E0,
		0.14,
		-0.14,
		true,
		false,
		null,
		"str",
		{
			"integer": 42,
			"negative_integer": -42,
			"float": 3.14,
			"negative_float": -3.14,
			"exponent_1": 314e-2,
			"exponent_2": 0.314e+1,
			"exponent_3": 3.14e0,
			"exponent_4": 314E-2,
			"exponent_5": 0.314E+1,
			"exponent_6": 3.14E0,
			"zero_beginning_float": 0.14,
			"negative_zero_beginning_float": -0.14,
			"boolean_true": true,
			"boolean_false": false,
			"nullable": null,
			"string": "str",
			"array": [
				42,
				-42,
				3.14,
				-3.14,
				314e-2,
				0.314e+1,
				3.14e0,
				314E-2,
				0.314E+1,
				3.14E0,
				0.14,
				-0.14,
				true,
				false,
				null,
				"str",
				{"object": {}}
			]
		}
	],
	"object": {
		"integer": 42,
		"negative_integer": -42,
		"float": 3.14,
		"negative_float": -3.14,
		"exponent_1": 314e-2,
		"exponent_2": 0.314e+1,
		"exponent_3": 3.14e0,
		"exponent_4": 314E-2,
		"exponent_5": 0.314E+1,
		"exponent_6": 3.14E0,
		"zero_beginning_float": 0.14,
		"negative_zero_beginning_float": -0.14,
		"boolean_true": true,
		"boolean_false": false,
		"nullable": null,
		"string": "str",
		"array": [
			42,
			-42,
			3.14,
			-3.14,
			314e-2,
			0.314e+1,
			3.14e0,
			314E-2,
			0.314E+1,
			3.14E0,
			0.14,
			-0.14,
			true,
			false,
			null,
			"str",
			{"object": {}}
		]
	}
}
`, 1797},
	"object_with_trailing_data": {`
{
	"foo": "bar"
}
with trailing data
`, 18},
}
