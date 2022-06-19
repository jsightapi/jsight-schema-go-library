package scanner

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/fs"
)

func TestScanner(t *testing.T) {
	cc := map[string][]operation{
		`  {  "key"  :  234  }  `: {
			assertLexVal("{"),
			assertLexVal(`"`), // opening quote
			assertLexVal(`"key"`),
			next(), // object value begin

			assertLexVal("2"), // first character of literal
			assertLexVal("234"),

			next(), // object value end

			assertLexVal(`{  "key"  :  234  }`),
		},

		`["str",false]`: {
			assertLexVal("["),

			next(), // array item begin

			assertLexVal(`"`), // opening quote
			assertLexVal(`"str"`),

			next(), // array item end
			next(), // array item begin

			assertLexVal("f"), // first character of the array item
			assertLexVal("false"),

			next(), // array item end

			assertLexVal(`["str",false]`),
		},

		"/* {} */": {
			assertLexVal("/*"),

			next(), // objBegin
			next(), // objEnd

			assertLexVal("/* {} */"),
		},

		"123 // { } - some comment\r\n": {
			assertLexVal("1"),
			assertLexVal("123"),
			assertLexVal("//"),
			assertLexVal("{"),
			assertLexVal("{ }"),
			assertLexVal("s"),
			assertLexVal("some comment"),
			assertLexVal("// { } - some comment"),
			assertLexVal("\r"),
			assertLexVal("\n"),
		},
	}

	for given, checkers := range cc {
		t.Run(given, func(t *testing.T) {
			file := fs.NewFile("", []byte(given))
			s := New(file)

			for _, c := range checkers {
				c.Check(t, s)
			}
		})
	}
}

type operation interface {
	Check(t *testing.T, s *Scanner)
}

type nextOperation struct{}

func next() nextOperation {
	return nextOperation{}
}

func (nextOperation) Check(t *testing.T, s *Scanner) {
	t.Helper()

	s.Next()
}

type assertLexValOperation struct {
	expected string
}

func assertLexVal(s string) assertLexValOperation {
	return assertLexValOperation{s}
}

func (c assertLexValOperation) Check(t *testing.T, s *Scanner) {
	t.Helper()

	lex, _ := s.Next()
	assert.Equal(t, c.expected, lex.Value().String())
}
