package json

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func BenchmarkIsNull(b *testing.B) {
	null := bytes.NewBytes(nullStr)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Guess(null).IsNull()
	}
}

var jsonTests = map[string][]string{
	`string`: {
		`""`,
		`"abc"`,
	},
	`integer`: {
		`0`,
		`-0`,
		`1`,
		`-1`,
		`1000`,
		`-1000`,
		`999999999999`,
		`-999999999999`,
		`0.0e1`,
		`-0.0e1`,
		`2e3`,
		`-2e3`,
		`2e+3`,
		`-2e+3`,
		`2.3e+1`,
		`-2.3e+1`,
		`2.34e+2`,
		`-2.24e+2`,
		`2.34e+20`,
		`-2.24e+20`,
	},
	`float`: {
		`0.1`,
		`-0.1`,
		`1.0`,
		`-1.0`,
		`2.345`,
		`-2.345`,
		`2.3e-4`,
		`-2.3e-4`,
		`2.3E-4`,
		`-2.3E-4`,
		`2.3e+0`,
		`-2.3e+0`,
		`2.34e+1`,
		`-2.34e+1`,
	},
	`boolean`: {
		`true`,
		`false`,
	},
	`null`: {
		`null`,
	},
	`wrong`: {
		``,
		`--1`,
		`1-`,
		`ABC`,
		`-ABC`,
		`1-2`,
		`3333-222-33`,
		`-`,
		`+`,
		`NULL`,
		`Null`,
		`TRUE`,
		`True`,
		`FALSE`,
		`False`,
		`'qwerty'`,
		`[]`,
		`{}`,
	},
}

// Returns all value from json variable for the specified key
func success(key string) []string {
	arr, ok := jsonTests[key]

	if !ok {
		panic(`Key "` + key + `" not found`)
	}

	return append([]string{}, arr...)
}

// Returns all the value from json variable with the exception of the specified key
func fail(key string) []string {
	var result []string
	for k, arr := range jsonTests {
		if k != key {
			result = append(result, arr...)
		}
	}
	return result
}

func TestIsInteger(t *testing.T) {
	for _, str := range success("integer") {
		t.Run(str, func(t *testing.T) {
			assert.True(t, Guess(bytes.NewBytes(str)).IsInteger())
		})
	}

	for _, str := range fail("integer") {
		t.Run(str, func(t *testing.T) {
			assert.False(t, Guess(bytes.NewBytes(str)).IsInteger())
		})
	}
}

func TestIsFloat(t *testing.T) {
	for _, str := range success("float") {
		t.Run(str, func(t *testing.T) {
			assert.True(t, Guess(bytes.NewBytes(str)).IsFloat())
		})
	}

	for _, str := range fail("float") {
		t.Run(str, func(t *testing.T) {
			assert.False(t, Guess(bytes.NewBytes(str)).IsFloat())
		})
	}
}

func TestIsString(t *testing.T) {
	for _, str := range success("string") {
		t.Run(str, func(t *testing.T) {
			assert.True(t, Guess(bytes.NewBytes(str)).IsString())
		})
	}

	for _, str := range fail("string") {
		t.Run(str, func(t *testing.T) {
			assert.False(t, Guess(bytes.NewBytes(str)).IsString())
		})
	}
}

func TestGuessData_IsObject(t *testing.T) {
	cc := map[string]bool{
		"{":  true,
		"":   false,
		" {": false,
		"{ ": false,
		"[":  false,
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			assert.Equal(t, expected, GuessData{bytes: bytes.NewBytes(given)}.IsObject())
		})
	}
}

func TestGuessData_IsArray(t *testing.T) {
	cc := map[string]bool{
		"[":  true,
		"":   false,
		" [": false,
		"[ ": false,
		"{":  false,
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			assert.Equal(t, expected, GuessData{bytes: bytes.NewBytes(given)}.IsArray())
		})
	}
}

func TestIsBoolean(t *testing.T) {
	for _, str := range success("boolean") {
		t.Run(str, func(t *testing.T) {
			assert.True(t, Guess(bytes.NewBytes(str)).IsBoolean())
		})
	}

	for _, str := range fail("boolean") {
		t.Run(str, func(t *testing.T) {
			assert.False(t, Guess(bytes.NewBytes(str)).IsBoolean())
		})
	}
}

func TestIsNull(t *testing.T) {
	for _, str := range success(nullStr) {
		t.Run(str, func(t *testing.T) {
			assert.True(t, Guess(bytes.NewBytes(str)).IsNull())
		})
	}

	for _, str := range fail(nullStr) {
		t.Run(str, func(t *testing.T) {
			assert.False(t, Guess(bytes.NewBytes(str)).IsNull())
		})
	}
}

func TestGuessLiteralNodeType(t *testing.T) {
	for _, str := range success("string") {
		t.Run(str, func(t *testing.T) {
			assert.Equal(t, TypeString, Guess(bytes.NewBytes(str)).LiteralJsonType())
		})
	}

	for _, str := range success("integer") {
		t.Run(str, func(t *testing.T) {
			assert.Equal(t, TypeInteger, Guess(bytes.NewBytes(str)).LiteralJsonType())
		})
	}

	for _, str := range success("float") {
		t.Run(str, func(t *testing.T) {
			assert.Equal(t, TypeFloat, Guess(bytes.NewBytes(str)).LiteralJsonType())
		})
	}

	for _, str := range success("boolean") {
		t.Run(str, func(t *testing.T) {
			assert.Equal(t, TypeBoolean, Guess(bytes.NewBytes(str)).LiteralJsonType())
		})
	}

	for _, str := range success(nullStr) {
		t.Run(str, func(t *testing.T) {
			assert.Equal(t, TypeNull, Guess(bytes.NewBytes(str)).LiteralJsonType())
		})
	}
}

func TestGuessLiteralNodeTypePanic(t *testing.T) {
	for _, str := range success("wrong") {
		assert.Panics(t, func() {
			Guess(bytes.NewBytes(str)).LiteralJsonType()
		})
	}
}

func TestNumberOptimization(t *testing.T) {
	b := bytes.NewBytes("123")
	g := Guess(b)

	g.IsInteger()
	pointer1 := g.number

	g.IsFloat()
	pointer2 := g.number

	pointer3, err := g.Number()
	require.NoError(t, err)

	pointer4 := g.number

	assert.NotNil(t, pointer1)

	assert.Equal(t, pointer1, pointer2)
	assert.Equal(t, pointer2, pointer3)
	assert.Equal(t, pointer3, pointer4)
}
