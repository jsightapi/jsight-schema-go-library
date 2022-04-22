package json

import (
	"fmt"
	"j/schema/bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkNewNumber(b *testing.B) {
	num := bytes.Bytes("-123.456E-5")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := NewNumber(num)
		if err != nil {
			panic(err)
		}
	}
}

func BenchmarkNewNumberFromInt(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewNumberFromInt(-123456)
	}
}

func BenchmarkNumber_Equal(b *testing.B) {
	number1, err := NewNumber(bytes.Bytes("-123456E-3"))
	require.NoError(b, err)
	number2, err := NewNumber(bytes.Bytes("-123.456E-3"))
	require.NoError(b, err)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		number1.Equal(number2)
	}
}

func TestNewNumber(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		var cc = []struct {
			jsonNumber string
			int        string
			fra        string
		}{
			{
				"111",
				"111",
				"",
			},
			{
				"222.1",
				"222",
				"1",
			},
			{
				"333.12",
				"333",
				"12",
			},
			{
				"444.123456789",
				"444",
				"123456789",
			},

			{
				"-555.1234567890000000000000",
				"555",
				"123456789",
			},

			{
				"-777.0",
				"777",
				"",
			},
			{
				"-888.00",
				"888",
				"",
			},

			{
				"-999.001",
				"999",
				"001",
			},

			{
				"-111e4", // --1110000
				"1110000",
				"",
			},
			{
				"-222E+4", // -2220000
				"2220000",
				"",
			},
			{
				"333e-1", // 33.3
				"33",
				"3",
			},
			{
				"444E-2", // 4.44
				"4",
				"44",
			},
			{
				"55.6e-2", // 0.556
				"",
				"556",
			},
			{
				"55.60e-2", // 0.5560
				"",
				"556",
			},
			{
				"55.6e-3", // 0.0556
				"",
				"0556",
			},
			{
				"55.67e-2", // 0.5567
				"",
				"5567",
			},
			{
				"55.678e-2", // 0.55678
				"",
				"55678",
			},
			{
				"123000e-5", // 1.23000,
				"1",
				"23",
			},
			{
				"100e-4", // 0.0100,
				"",
				"01",
			},
			{
				"10.010e-2", // 0.10010,
				"",
				"1001",
			},
			{
				"1010.00e-2", // 10.1000,
				"10",
				"1",
			},

			{
				"100e-2",
				"1",
				"",
			},
			{
				"0.1e+1",
				"1",
				"",
			},

			{
				"20",
				"20",
				"",
			},
			{
				"0.200e2",
				"20",
				"",
			},
			{
				"0.20e2",
				"20",
				"",
			},
			{
				"0.2e2",
				"20",
				"",
			},
			{
				"2e1",
				"20",
				"",
			},
			{
				"20e0",
				"20",
				"",
			},
			{
				"200e-1",
				"20",
				"",
			},
			{
				"2000e-2",
				"20",
				"",
			},

			{
				"0.0123E4",
				"123",
				"",
			},
			{
				"0.0001200e+2",
				"",
				"012",
			},

			{
				"0.001200e+2",
				"",
				"12",
			},
			{
				"0.00120e+2",
				"",
				"12",
			},
			{
				"0.0012e+2",
				"",
				"12",
			},
			{
				"0.012e+1",
				"",
				"12",
			},
			{
				"0.12e0",
				"",
				"12",
			},
			{
				"0.12",
				"",
				"12",
			},
			{
				"1.2e-1",
				"",
				"12",
			},
			{
				"12e-2",
				"",
				"12",
			},
			{
				"12.0e-2",
				"",
				"12",
			},
			{
				"12.00e-2",
				"",
				"12",
			},
			{
				"12.00e-2",
				"",
				"12",
			},

			{
				"0.0e0",
				"",
				"",
			},
		}

		for _, c := range cc {
			t.Run(c.jsonNumber, func(t *testing.T) {
				number, err := NewNumber(bytes.Bytes(c.jsonNumber))
				require.NoError(t, err)

				assert.Equal(t, c.int, string(number.int()))
				assert.Equal(t, c.fra, string(number.fra()))
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		var ss = []string{ // http://json.org/
			"",
			"01",
			"0e0",
			"0e2",
			"+1",
			".1",
			"-.1",
			"-",
			"e2",
			"E2",
			"1.1e2e",
			"1.e2",
			`"abc"`,
			"abc",
		}

		for _, s := range ss {
			t.Run(s, func(t *testing.T) {
				_, err := NewNumber(bytes.Bytes(s))
				assert.Error(t, err)
			})
		}
	})
}

func TestNumber_trimLeadingZerosInTheIntegerPart(t *testing.T) {
	cc := []struct {
		number   Number
		expected string
	}{
		{Number{false, bytes.Bytes("00123"), 2}, "123"},
		{Number{false, bytes.Bytes("0123"), 2}, "123"},
		{Number{false, bytes.Bytes("123"), 2}, "123"},
		{Number{false, bytes.Bytes("023"), 2}, "23"},
		{Number{false, bytes.Bytes("0023"), 3}, "023"},
		{Number{false, bytes.Bytes("00023"), 3}, "023"},
		{Number{false, bytes.Bytes("00023"), 4}, "0023"},
	}

	for _, c := range cc {
		t.Run(c.number.String(), func(t *testing.T) {
			err := c.number.trimLeadingZerosInTheIntegerPart()
			require.NoError(t, err)
			assert.Equal(t, c.expected, string(c.number.nat))
		})
	}
}

func TestNumber_trimTrailingZerosInTheFractionalPart(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := []struct {
			number   Number
			expected string
		}{
			{Number{false, bytes.Bytes("123000"), 0}, "123000"},
			{Number{false, bytes.Bytes("123000"), 1}, "12300"},
			{Number{false, bytes.Bytes("123000"), 2}, "1230"},
			{Number{false, bytes.Bytes("123000"), 3}, "123"},
			{Number{false, bytes.Bytes("123000"), 4}, "123"},
			{Number{false, bytes.Bytes("123000"), 5}, "123"},
		}

		for _, c := range cc {
			t.Run(c.number.String(), func(t *testing.T) {
				err := c.number.trimTrailingZerosInTheFractionalPart()
				require.NoError(t, err)
				assert.Equal(t, c.expected, string(c.number.nat))
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]Number{
			"negative exponent":                   {false, bytes.Bytes("000"), -1},
			"exponent greater than length of nat": {false, bytes.Bytes("000"), 4},
		}

		for name, n := range cc {
			t.Run(name, func(t *testing.T) {
				err := n.trimTrailingZerosInTheFractionalPart()
				assert.EqualError(t, err, "incorrect exponent value")
			})
		}
	})
}

func TestNumber_LengthOfFractionalPart(t *testing.T) {
	cc := map[string]uint{
		"123":     0,
		"123.4":   1,
		"123.45":  2,
		"0.123":   3,
		"123e2":   0,
		"123e-2":  2,
		"123e-4":  4,
		"1.23e-4": 6,
		"-0.123":  3,
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			n, err := NewNumber(bytes.Bytes(given))
			require.NoError(t, err)
			assert.Equal(t, expected, n.LengthOfFractionalPart())
		})
	}
}

func TestNewNumberFromInt(t *testing.T) {
	cc := map[int]string{
		-12: "-12",
		-1:  "-1",
		0:   "0",
		1:   "1",
		12:  "12",
		123: "123",
	}

	for given, expected := range cc {
		t.Run(expected, func(t *testing.T) {
			actual := NewNumberFromInt(given).String()
			assert.Equal(t, expected, actual)
		})
	}
}

func TestNumber_Cmp(t *testing.T) {
	var cc = []struct {
		number1  string
		number2  string
		expected int
	}{
		// equal
		{
			"123",
			"123",
			0,
		},
		{
			"123.45",
			"123.45",
			0,
		},

		{
			"123.4560",
			"123.456",
			0,
		},
		{
			"123.456",
			"123.456000",
			0,
		},

		{
			"123.001",
			"123001e-3",
			0,
		},
		{
			"123E+1",
			"12300E-1",
			0,
		},

		// integer
		{
			"0",
			"0",
			0,
		},
		{
			"0",
			"1",
			-1,
		},
		{
			"1",
			"0",
			1,
		},
		{
			"111",
			"222",
			-1,
		},
		{
			"222",
			"111",
			1,
		},
		{
			"1111",
			"111",
			1,
		},
		{
			"111",
			"1111",
			-1,
		},
		{
			"16",
			"25",
			-1,
		},

		// negative
		{
			"-123",
			"123",
			-1,
		},
		{
			"123",
			"-123",
			1,
		},
		{
			"-123.45",
			"123.45",
			-1,
		},
		{
			"123.45",
			"-123.45",
			1,
		},

		// fractional
		{
			"123.001",
			"123.002",
			-1,
		},
		{
			"123.002",
			"123.001",
			1,
		},

		{
			"123.456",
			"123.4567",
			-1,
		},
		{
			"123.4567",
			"123.456",
			1,
		},
		{
			"-1",
			"-2",
			1,
		},
		{
			"-2",
			"-1",
			-1,
		},
		{
			"-1",
			"-1",
			0,
		},
	}

	for _, c := range cc {
		t.Run(fmt.Sprintf("%s %s", c.number1, c.number2), func(t *testing.T) {
			number1, err := NewNumber(bytes.Bytes(c.number1))
			require.NoError(t, err)
			number2, err := NewNumber(bytes.Bytes(c.number2))
			require.NoError(t, err)

			assert.Equal(t, c.expected, number1.Cmp(number2))
		})
	}
}

func TestNumber_Equal(t *testing.T) {
	var cc = []struct {
		number1  string
		number2  string
		expected bool
	}{
		// equal
		{
			"123",
			"123",
			true,
		},
		{
			"11",
			"22",
			false,
		},
		{
			"22",
			"11",
			false,
		},
		{
			"0",
			"0",
			true,
		},
		{
			"-1",
			"-1",
			true,
		},
	}

	for _, c := range cc {
		t.Run(fmt.Sprintf("%s == %s", c.number1, c.number2), func(t *testing.T) {
			number1, err := NewNumber(bytes.Bytes(c.number1))
			require.NoError(t, err)

			number2, err := NewNumber(bytes.Bytes(c.number2))
			require.NoError(t, err)

			assert.Equal(t, c.expected, number1.Equal(number2))
		})
	}
}

func TestNumber_GreaterThan(t *testing.T) {
	var cc = []struct {
		number1  string
		number2  string
		expected bool
	}{
		// equal
		{
			"0",
			"0",
			false,
		},
		{
			"0",
			"1",
			false,
		},
		{
			"1",
			"0",
			true,
		},
		{
			"123",
			"123",
			false,
		},
		{
			"16",
			"25",
			false,
		},
		{
			"25",
			"16",
			true,
		},
		{
			"-1",
			"-1",
			false,
		},
		{
			"-1",
			"-2",
			true,
		},
		{
			"-2",
			"-1",
			false,
		},
	}

	for _, c := range cc {
		t.Run(fmt.Sprintf("%s > %s", c.number1, c.number2), func(t *testing.T) {
			number1, err := NewNumber(bytes.Bytes(c.number1))
			require.NoError(t, err)

			number2, err := NewNumber(bytes.Bytes(c.number2))
			require.NoError(t, err)

			assert.Equal(t, c.expected, number1.GreaterThan(number2))
		})
	}
}

func TestNumber_GreaterThanOrEqual(t *testing.T) {
	var cc = []struct {
		number1  string
		number2  string
		expected bool
	}{
		// equal
		{
			"0",
			"0",
			true,
		},
		{
			"0",
			"1",
			false,
		},
		{
			"1",
			"0",
			true,
		},
		{
			"123",
			"123",
			true,
		},
		{
			"16",
			"25",
			false,
		},
		{
			"25",
			"16",
			true,
		},
		{
			"-1",
			"-1",
			true,
		},
		{
			"-1",
			"-2",
			true,
		},
		{
			"-2",
			"-1",
			false,
		},
	}

	for _, c := range cc {
		t.Run(fmt.Sprintf("%s >= %s", c.number1, c.number2), func(t *testing.T) {
			number1, err := NewNumber(bytes.Bytes(c.number1))
			require.NoError(t, err)

			number2, err := NewNumber(bytes.Bytes(c.number2))
			require.NoError(t, err)

			assert.Equal(t, c.expected, number1.GreaterThanOrEqual(number2))
		})
	}
}

func TestNumber_LessThan(t *testing.T) {
	var cc = []struct {
		number1  string
		number2  string
		expected bool
	}{
		// equal
		{
			"0",
			"0",
			false,
		},
		{
			"1",
			"0",
			false,
		},
		{
			"0",
			"1",
			true,
		},
		{
			"123",
			"123",
			false,
		},
		{
			"16",
			"25",
			true,
		},
		{
			"25",
			"16",
			false,
		},
		{
			"-1",
			"-1",
			false,
		},
		{
			"-1",
			"-2",
			false,
		},
		{
			"-2",
			"-1",
			true,
		},
	}

	for _, c := range cc {
		t.Run(fmt.Sprintf("%s < %s", c.number1, c.number2), func(t *testing.T) {
			number1, err := NewNumber(bytes.Bytes(c.number1))
			require.NoError(t, err)

			number2, err := NewNumber(bytes.Bytes(c.number2))
			require.NoError(t, err)

			assert.Equal(t, c.expected, number1.LessThan(number2))
		})
	}
}

func TestNumber_LessThanOrEqual(t *testing.T) {
	var cc = []struct {
		number1  string
		number2  string
		expected bool
	}{
		// equal
		{
			"0",
			"0",
			true,
		},
		{
			"1",
			"0",
			false,
		},
		{
			"0",
			"1",
			true,
		},
		{
			"123",
			"123",
			true,
		},
		{
			"16",
			"25",
			true,
		},
		{
			"25",
			"16",
			false,
		},
		{
			"-1",
			"-1",
			true,
		},
		{
			"-1",
			"-2",
			false,
		},
		{
			"-2",
			"-1",
			true,
		},
	}

	for _, c := range cc {
		t.Run(fmt.Sprintf("%s <= %s", c.number1, c.number2), func(t *testing.T) {
			number1, err := NewNumber(bytes.Bytes(c.number1))
			require.NoError(t, err)

			number2, err := NewNumber(bytes.Bytes(c.number2))
			require.NoError(t, err)

			assert.Equal(t, c.expected, number1.LessThanOrEqual(number2))
		})
	}
}

func TestNumber_ToFloat(t *testing.T) {
	cc := map[string]float64{
		"42":   42,
		"3.14": 3.14,
		"2e3":  2e3,
		"2e-3": 2e-3,
	}

	for number, expected := range cc {
		t.Run(number, func(t *testing.T) {
			n, err := NewNumber([]byte(number))
			require.NoError(t, err)
			assert.Equal(t, expected, n.ToFloat())
		})
	}
}
