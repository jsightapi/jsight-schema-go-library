package bytes

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBytes_Equals(t *testing.T) {
	const given = "foo"
	cc := map[string]bool{
		given: true,
		"":    false,
		"fOo": false,
		"bar": false,
	}

	for bb, expected := range cc {
		t.Run(bb, func(t *testing.T) {
			actual := Bytes(given).Equals(Bytes(bb))
			assert.Equal(t, expected, actual)
		})
	}
}

func TestBytes_Slice(t *testing.T) {
	actual := Bytes("1234567890").Slice(2, 6)
	assert.Equal(t, "34567", string(actual))
}

func TestBytes_Unquote(t *testing.T) {
	cc := map[string]string{
		// trimmed
		`""`:    "",
		`"123"`: `123`,
		`"\\n"`: `\n`,

		// no trimmed
		"":     "",
		`"`:    `"`,
		"123":  "123",
		`"123`: `"123`,
		`123"`: `123"`,

		`"\"\u0061bc\""`: `"abc"`,
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			actual := Bytes(given).Unquote()
			assert.Equal(t, expected, string(actual))
		})
	}
}

func TestBytes_TrimSquareBrackets(t *testing.T) {
	cc := map[string]string{
		"":        "",
		"foo":     "foo",
		"[foo":    "[foo",
		"foo]":    "foo]",
		"[foo]":   "foo",
		"{foo}":   "{foo}",
		"(foo)":   "(foo)",
		"[[foo]]": "[foo]",
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			actual := Bytes(given).TrimSquareBrackets()
			assert.Equal(t, expected, string(actual))
		})
	}
}

func TestBytes_TrimSpaces(t *testing.T) {
	cc := map[string]string{
		"":              "",
		" \t \n\r\r   ": "",
		"1":             "1",
		"12":            "12",

		" 123":              "123",
		"\t123":             "123",
		"\n123":             "123",
		"\r123":             "123",
		"\t\t\n\n\r\r  123": "123",

		"123 ":             "123",
		"123\t":            "123",
		"123\n":            "123",
		"123\r":            "123",
		"123\t\t\n\n\n\n ": "123",

		" 123 ":                           "123",
		"\t123\t":                         "123",
		"\n123\n":                         "123",
		"\r123\r":                         "123",
		"\t\t\n\n\r\r  123\t\t\n\n\n\n  ": "123",
		"\t123\t\t\n\n\n\n  ":             "123",
		"\t\t\n\n\r\r  123\t":             "123",
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			actual := Bytes(given).TrimSpaces()
			assert.Equal(t, expected, string(actual))
		})
	}
}

func TestBytes_TrimSpacesFromLeft(t *testing.T) {
	cc := map[string]string{
		"":    "",
		"1":   "1",
		"12":  "12",
		"123": "123",

		" 123":              "123",
		"\t123":             "123",
		"\n123":             "123",
		"\r123":             "123",
		"\t\t\n\n\r\r  123": "123",

		"123 ":  "123 ",
		"123\t": "123\t",
		"123\n": "123\n",
		"123\r": "123\r",
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			actual := Bytes(given).TrimSpacesFromLeft()
			assert.Equal(t, expected, string(actual))
		})
	}
}

func TestBytes_CountSpacesFromLeft(t *testing.T) {
	cc := map[string]int{
		"":           0,
		"foo":        0,
		" \t\r\nfoo": 4,
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			actual := Bytes(given).CountSpacesFromLeft()
			assert.Equal(t, expected, actual)
		})
	}
}

func TestBytes_OneOf(t *testing.T) {
	b := Bytes("foo")

	cc := []struct {
		given    []string
		expected bool
	}{
		{[]string{"foo", "bar", "fizz", "buzz"}, true},
		{[]string{"buzz", "bar", "fizz", "foo"}, true},
		{[]string{"buzz", "foo", "fizz", "bar"}, true},
		{[]string{"foo", "foo"}, true},
		{nil, false},
		{[]string{}, false},
		{[]string{"bar"}, false},
		{[]string{" foo", "Foo"}, false},
	}

	for _, c := range cc {
		t.Run(fmt.Sprintf("%v", c.given), func(t *testing.T) {
			actual := b.OneOf(c.given...)
			assert.Equal(t, c.expected, actual)
		})
	}
}

func BenchmarkBytes_OneOf(b *testing.B) {
	pp := []string{"foo", "bar", "fizz", "buzz"}

	bytes := Bytes("buzz")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bytes.OneOf(pp...)
	}
}

func TestBytes_ParseBool(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]bool{
			"true":  true,
			"false": false,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				actual, err := Bytes(given).ParseBool()
				require.NoError(t, err)
				assert.Equal(t, expected, actual)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		ss := []string{
			"",
			"True",
			"fAlSe",
			"foo",
		}

		for _, s := range ss {
			t.Run(s, func(t *testing.T) {
				_, err := Bytes(s).ParseBool()
				assert.EqualError(t, err, "invalid bool value")
			})
		}
	})
}

var benchmarkParseIntBytes = Bytes("1234567890")

func TestBytes_ParseUint(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]uint{
			"42":    42,
			"00000": 0,
			"00042": 42,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				actual, err := Bytes(given).ParseUint()

				require.NoError(t, err)
				assert.Equal(t, expected, actual)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]string{
			"":     "not enough data in ParseUint",
			"3.14": "invalid byte (.) found in ParseUint (3.14)",
			"-1":   "invalid byte (-) found in ParseUint (-1)",
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				_, err := Bytes(given).ParseUint()

				assert.EqualError(t, err, expected)
			})
		}
	})
}

func BenchmarkParseUint(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := benchmarkParseIntBytes.ParseUint()
		assert.NoError(b, err)
	}
}

func TestBytes_ParseInt(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]int{
			"3":                             3,
			"23":                            23,
			"123":                           123,
			"00123":                         123,
			"-123":                          -123,
			strconv.Itoa(math.MaxInt):       math.MaxInt,
			"-" + strconv.Itoa(math.MaxInt): -math.MaxInt,
		}
		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				actual, err := Bytes(given).ParseInt()
				require.NoError(t, err)
				assert.Equal(t, expected, actual)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]string{
			"3.2":                           "invalid byte (.) found in ParseUint (3.2)",
			strconv.Itoa(math.MaxInt) + "0": "too much data for int",
		}
		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				_, err := Bytes(given).ParseInt()
				assert.EqualError(t, err, expected)
			})
		}
	})
}

func BenchmarkParseInt(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := benchmarkParseIntBytes.ParseInt()
		assert.NoError(b, err)
	}
}

func TestBytes_IsUserTypeName(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		var tests = []string{
			"@-",
			"@_",
			"@ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			"@abcdefghijklmnopqrstuvwxyz",
			"@0123456789",
			"@ABCDEFGHIJKLMNOPQRSTUVWXYZ-abcdefghijklmnopqrstuvwxyz_0123456789",
			"@abc-",
		}

		for _, str := range tests {
			t.Run(str, func(t *testing.T) {
				assert.True(t, Bytes(str).IsUserTypeName())
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		var tests = []string{
			"",
			"@",
			"-",
			"_",
			"ABC",
			"@a.b",
			"-@abc",
			"@@",
		}

		for _, str := range tests {
			t.Run(str, func(t *testing.T) {
				assert.False(t, Bytes(str).IsUserTypeName())
			})
		}
	})
}

func TestBytes_String(t *testing.T) {
	cc := map[string]Bytes{
		"foo":          []byte("foo"),
		"\u0001\u0002": []byte{1, 2},
	}

	for expected, given := range cc {
		t.Run(expected, func(t *testing.T) {
			assert.Equal(t, expected, given.String())
		})
	}
}

func TestBytes_Len(t *testing.T) {
	cc := map[string]int{
		"":    0,
		"foo": 3,
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			actual := Bytes(given).Len()
			assert.Equal(t, expected, actual)
		})
	}
}

func TestBytes_LineFrom(t *testing.T) {
	tests := []struct {
		b       string
		s       int
		want    string
		wantErr bool
	}{
		{
			"abc",
			-1,
			"",
			true,
		},
		{
			"abc",
			0,
			"abc",
			false,
		},
		{
			"abc",
			1,
			"bc",
			false,
		},
		{
			"abc",
			2,
			"c",
			false,
		},
		{
			"abc",
			3,
			"",
			false,
		},
		{
			"abc",
			4,
			"",
			true,
		},
		{
			"abc\n123\nxyz",
			0,
			"abc",
			false,
		},
		{
			"abc\n123\nxyz",
			1,
			"bc",
			false,
		},
		{
			"abc\n123\nxyz",
			2,
			"c",
			false,
		},
		{
			"abc\n123\nxyz",
			3,
			"",
			false,
		},
		{
			"abc\n123\nxyz",
			4,
			"123",
			false,
		},
		{
			"abc\n123\nxyz",
			5,
			"23",
			false,
		},
		{
			"abc\n123\nxyz",
			6,
			"3",
			false,
		},
		{
			"abc\n123\nxyz",
			7,
			"",
			false,
		},
		{
			"abc\n123\nxyz",
			8,
			"xyz",
			false,
		},
		{
			"abc\n123\nxyz",
			9,
			"yz",
			false,
		},
		{
			"abc\n123\nxyz",
			9999,
			"",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.b, func(t *testing.T) {
			b := Bytes(tt.b)
			got, err := b.LineFrom(Index(tt.s))
			if tt.wantErr {
				require.NotNilf(t, err, "b.LineFrom(%v)", tt.s)
			} else {
				require.NoErrorf(t, err, "b.LineFrom(%v)", tt.s)
				require.Equalf(t, Bytes(tt.want), got, "b.LineFrom(%v)", tt.s)
			}
		})
	}
}

func TestBytes_Normalize(t *testing.T) {
	cc := map[string]string{
		"foo":          "foo",
		"\u0061":       "a",
		`\u0061`:       "a",
		`\uD83E\uDD10`: "🤐",
		"foo bar":      "foo bar",
		`
\u0061
foo bar
	\uD83E\uDD10 smile
fizz	buzz
`: `
a
foo bar
	🤐 smile
fizz	buzz
`,
		`\\u0061`:  `\\u0061`,
		`\\\u0061`: `\\a`,
		`\uffff`:   `￿`,
		`\b`:       "\b",
		`\f`:       "\f",
		`\n`:       "\n",
		`\r`:       "\r",
		`\t`:       "\t",
		`"\t\u0063"`: `"	c"`,
		`\\`:             `\\`,
		`"\"\u0061bc\""`: `"\"abc\""`,
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			actual := Bytes(given).Normalize()
			assert.Equal(t, expected, string(actual))
		})
	}
}
