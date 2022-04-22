package bytes

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			if actual != c.expected {
				t.Errorf("%t != %t", actual, c.expected)
			}
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

type bytesTestData struct {
	source string
	result string
}

var benchmarkParseIntBytes = Bytes("1234567890")

func BenchmarkParseUint(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := benchmarkParseIntBytes.ParseUint()
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkParseInt(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := benchmarkParseIntBytes.ParseInt()
		if err != nil {
			b.Error(err)
		}
	}
}

func BenchmarkAtoi(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := strconv.Atoi(string(benchmarkParseIntBytes))
		if err != nil {
			b.Error(err)
		}
	}
}

func TestBytes_Unquote(t *testing.T) {
	testData := []bytesTestData{
		// trimmed
		{`""`, ``},
		{`"123"`, `123`},

		// no trimmed
		{``, ``},
		{`"`, `"`},
		{`123`, `123`},
		{`"123`, `"123`},
		{`123"`, `123"`},
	}

	for _, d := range testData {
		source := Bytes(d.source)
		trimmed := string(source.Unquote())
		if trimmed != d.result {
			t.Errorf(`Incorrect result %#v for source %#v expected %#v`, trimmed, d.source, d.result)
		}
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
	testData := []bytesTestData{
		{"", ""},
		{"1", "1"},
		{"12", "12"},
		{"123", "123"},

		{" 123", "123"},
		{"\t123", "123"},
		{"\n123", "123"},
		{"\r123", "123"},
		{"\t\t\n\n\r\r  123", "123"},

		{"123 ", `123 `},
		{"123\t", "123\t"},
		{"123\n", "123\n"},
		{"123\r", "123\r"},
	}

	for _, d := range testData {
		source := Bytes(d.source)
		trimmed := string(source.TrimSpacesFromLeft())
		if trimmed != d.result {
			t.Errorf(`Incorrect result %#v for source %#v expected %#v`, trimmed, d.source, d.result)
		}
	}
}

func TestBytes_ParseInt(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]int{
			"3":                        3,
			"23":                       23,
			"123":                      123,
			"00123":                    123,
			"-123":                     -123,
			strconv.Itoa(maxInt):       maxInt,
			"-" + strconv.Itoa(maxInt): -maxInt,
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
			"3.2":                      "invalid byte (.) found in ParseUint (3.2)",
			strconv.Itoa(maxInt) + "0": "too much data for int",
		}
		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				_, err := Bytes(given).ParseInt()
				assert.EqualError(t, err, expected)
			})
		}
	})
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

func TestIsValidSchemaNameByte(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		// Allowed special symbols.
		cc := []byte{
			'-',
			'_',
		}

		// Any english lower letters.
		for i := 'a'; i <= 'z'; i++ {
			cc = append(cc, byte(i))
		}

		// Any english capital letters.
		for i := 'A'; i <= 'Z'; i++ {
			cc = append(cc, byte(i))
		}

		// Any digits.
		for i := '0'; i <= '9'; i++ {
			cc = append(cc, byte(i))
		}

		for _, c := range cc {
			t.Run(string(c), func(t *testing.T) {
				assert.True(t, IsValidUserTypeNameByte(c))
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := make([]byte, 0, 255)

		for i := 0; i <= 255; i++ {
			if i == '-' || i == '_' || ('a' <= i && i <= 'z') || ('A' <= i && i <= 'Z') || ('0' <= i && i <= '9') {
				continue
			}
			cc = append(cc, byte(i))
		}

		for _, c := range cc {
			t.Run(string(c), func(t *testing.T) {
				assert.False(t, IsValidUserTypeNameByte(c))
			})
		}
	})
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
