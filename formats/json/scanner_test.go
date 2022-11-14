package json

import (
	"reflect"
	"testing"

	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/lexeme"

	"github.com/stretchr/testify/assert"
)

func TestScanner_Next(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		jsonValidResults := map[string][]lexeme.LexEventType{
			"12.34":               {lexeme.LiteralBegin, lexeme.LiteralEnd},
			" 12.34 ":             {lexeme.LiteralBegin, lexeme.LiteralEnd},
			"12.34\n":             {lexeme.LiteralBegin, lexeme.LiteralEnd},
			"12.34\r\n":           {lexeme.LiteralBegin, lexeme.LiteralEnd},
			"12.34 ":              {lexeme.LiteralBegin, lexeme.LiteralEnd},
			"12.34 \r\n":          {lexeme.LiteralBegin, lexeme.LiteralEnd},
			`"str"`:               {lexeme.LiteralBegin, lexeme.LiteralEnd},
			`"str" `:              {lexeme.LiteralBegin, lexeme.LiteralEnd},
			`"\u0000"`:            {lexeme.LiteralBegin, lexeme.LiteralEnd},
			`"\\" `:               {lexeme.LiteralBegin, lexeme.LiteralEnd},
			"true":                {lexeme.LiteralBegin, lexeme.LiteralEnd},
			"false":               {lexeme.LiteralBegin, lexeme.LiteralEnd},
			"null":                {lexeme.LiteralBegin, lexeme.LiteralEnd},
			"-1":                  {lexeme.LiteralBegin, lexeme.LiteralEnd},
			"0.123":               {lexeme.LiteralBegin, lexeme.LiteralEnd},
			"-0.123":              {lexeme.LiteralBegin, lexeme.LiteralEnd},
			"1e2":                 {lexeme.LiteralBegin, lexeme.LiteralEnd},
			"1.23e+11":            {lexeme.LiteralBegin, lexeme.LiteralEnd},
			"[]":                  {lexeme.ArrayBegin, lexeme.ArrayEnd},
			"[ ]":                 {lexeme.ArrayBegin, lexeme.ArrayEnd},
			"[1 ]":                {lexeme.ArrayBegin, lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd, lexeme.ArrayEnd},
			"[{}]":                {lexeme.ArrayBegin, lexeme.ArrayItemBegin, lexeme.ObjectBegin, lexeme.ObjectEnd, lexeme.ArrayItemEnd, lexeme.ArrayEnd},
			"[[]]":                {lexeme.ArrayBegin, lexeme.ArrayItemBegin, lexeme.ArrayBegin, lexeme.ArrayEnd, lexeme.ArrayItemEnd, lexeme.ArrayEnd},
			"{}":                  {lexeme.ObjectBegin, lexeme.ObjectEnd},
			"{} ":                 {lexeme.ObjectBegin, lexeme.ObjectEnd},
			" {} ":                {lexeme.ObjectBegin, lexeme.ObjectEnd},
			`{"foo":"bar"}`:       {lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.ObjectEnd},
			` { "foo" : "bar" } `: {lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd, lexeme.ObjectEnd},
			`["",[]]`: {
				lexeme.ArrayBegin,
				lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
				lexeme.ArrayItemBegin, lexeme.ArrayBegin, lexeme.ArrayEnd, lexeme.ArrayItemEnd,
				lexeme.ArrayEnd,
			},
			`{"foo": "bar", "key": 1}`: {
				lexeme.ObjectBegin,
				lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd,
				lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd,
				lexeme.ObjectEnd,
			},
			`[1,"str",false]`: {
				lexeme.ArrayBegin,
				lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
				lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
				lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
				lexeme.ArrayEnd,
			},
			`{"foo": [1,"str",false]}`: {
				lexeme.ObjectBegin,
				lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd,
				lexeme.ObjectValueBegin,
				lexeme.ArrayBegin,
				lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
				lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
				lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
				lexeme.ArrayEnd,
				lexeme.ObjectValueEnd,
				lexeme.ObjectEnd,
			},
			`{
	
		"foo"
	
		:
	
		123
	
		}`: {
				lexeme.ObjectBegin,
				lexeme.ObjectKeyBegin,
				lexeme.ObjectKeyEnd,
				lexeme.ObjectValueBegin,
				lexeme.LiteralBegin,
				lexeme.LiteralEnd,
				lexeme.ObjectValueEnd,
				lexeme.ObjectEnd,
			},
			`
		{
			"a": 1,
			"b": [2,3,4],
			"c": 5
		}`: {
				lexeme.ObjectBegin,
				lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd,
				lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd,
				lexeme.ObjectValueBegin,
				lexeme.ArrayBegin,
				lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
				lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
				lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
				lexeme.ArrayEnd,
				lexeme.ObjectValueEnd,
				lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd,
				lexeme.ObjectEnd,
			},
			`
		[
			1,
			{"k": 2},
			3
		]`: {
				lexeme.ArrayBegin,
				lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
				lexeme.ArrayItemBegin,
				lexeme.ObjectBegin,
				lexeme.ObjectKeyBegin, lexeme.ObjectKeyEnd, lexeme.ObjectValueBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ObjectValueEnd,
				lexeme.ObjectEnd,
				lexeme.ArrayItemEnd,
				lexeme.ArrayItemBegin, lexeme.LiteralBegin, lexeme.LiteralEnd, lexeme.ArrayItemEnd,
				lexeme.ArrayEnd,
			},
		}

		for json, expected := range jsonValidResults {
			t.Run("json", func(t *testing.T) {
				s := newScanner(fs.NewFile("", json))
				var results []lexeme.LexEventType

				for {
					if lex, ok := s.Next(); ok {
						results = append(results, lex.Type())
					} else {
						break
					}
				}

				assert.Truef(t, reflect.DeepEqual(results, expected), "Wrong results:\n%s\n\n%v", json, results)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]string{
			"+1": `ERROR (code 301): Invalid character "+" looking for beginning of value
	in line 1 on file 
	> +1
	--^`,
			"zzz": `ERROR (code 301): Invalid character "z" looking for beginning of value
	in line 1 on file 
	> zzz
	--^`,
			"tRue": `ERROR (code 301): Invalid character "R" in literal true (expecting 'r')
	in line 1 on file 
	> tRue
	---^`,
			"trUe": `ERROR (code 301): Invalid character "U" in literal true (expecting 'u')
	in line 1 on file 
	> trUe
	----^`,
			"truE": `ERROR (code 301): Invalid character "E" in literal true (expecting 'e')
	in line 1 on file 
	> truE
	-----^`,
			"tru": `ERROR (code 303): Unexpected end of file
	in line 1 on file 
	> tru
	----^`,
			"fAlse": `ERROR (code 301): Invalid character "A" in literal false (expecting 'a')
	in line 1 on file 
	> fAlse
	---^`,
			"faLse": `ERROR (code 301): Invalid character "L" in literal false (expecting 'l')
	in line 1 on file 
	> faLse
	----^`,
			"falSe": `ERROR (code 301): Invalid character "S" in literal false (expecting 's')
	in line 1 on file 
	> falSe
	-----^`,
			"falsE": `ERROR (code 301): Invalid character "E" in literal false (expecting 'e')
	in line 1 on file 
	> falsE
	------^`,
			"fal": `ERROR (code 303): Unexpected end of file
	in line 1 on file 
	> fal
	----^`,
			"nUll": `ERROR (code 301): Invalid character "U" in literal null (expecting 'u')
	in line 1 on file 
	> nUll
	---^`,
			"nuLl": `ERROR (code 301): Invalid character "L" in literal null (expecting 'l')
	in line 1 on file 
	> nuLl
	----^`,
			"nulL": `ERROR (code 301): Invalid character "L" in literal null (expecting 'l')
	in line 1 on file 
	> nulL
	-----^`,
			"nul": `ERROR (code 303): Unexpected end of file
	in line 1 on file 
	> nul
	----^`,
			`"	"`: `ERROR (code 301): Invalid character "\t" in string literal
	in line 1 on file 
	> "	"
	---^`,
			`"\x"`: `ERROR (code 301): Invalid character "x" in string escape code
	in line 1 on file 
	> "\x"
	----^`,
			`"\uZ"`: `ERROR (code 301): Invalid character "Z" in \u hexadecimal character escape
	in line 1 on file 
	> "\uZ"
	-----^`,
			`"\u1Z"`: `ERROR (code 301): Invalid character "Z" in \u hexadecimal character escape
	in line 1 on file 
	> "\u1Z"
	------^`,
			`"\u22Z"`: `ERROR (code 301): Invalid character "Z" in \u hexadecimal character escape
	in line 1 on file 
	> "\u22Z"
	-------^`,
			`"\u33Z"`: `ERROR (code 301): Invalid character "Z" in \u hexadecimal character escape
	in line 1 on file 
	> "\u33Z"
	-------^`,
			`"\u444Z"`: `ERROR (code 301): Invalid character "Z" in \u hexadecimal character escape
	in line 1 on file 
	> "\u444Z"
	--------^`,
			"-z": `ERROR (code 301): Invalid character "z" in numeric literal
	in line 1 on file 
	> -z
	---^`,
			"5.1.2": `ERROR (code 301): Invalid character "." non-space byte after top-level value
	in line 1 on file 
	> 5.1.2
	-----^`,
			`2"`: `ERROR (code 301): Invalid character "\"" non-space byte after top-level value
	in line 1 on file 
	> 2"
	---^`,
			"2'": `ERROR (code 301): Invalid character "'" non-space byte after top-level value
	in line 1 on file 
	> 2'
	---^`,
			"0.z": `ERROR (code 301): Invalid character "z" after decimal point in numeric literal
	in line 1 on file 
	> 0.z
	----^`,
			"1.23e+Z": `ERROR (code 301): Invalid character "Z" in exponent of numeric literal
	in line 1 on file 
	> 1.23e+Z
	--------^`,
			"[}": `ERROR (code 301): Invalid character "}" looking for beginning of value
	in line 1 on file 
	> [}
	---^`,
			"[1,]": `ERROR (code 301): Invalid character "]" looking for beginning of value
	in line 1 on file 
	> [1,]
	-----^`,
			"[1:]": `ERROR (code 301): Invalid character ":" after array item
	in line 1 on file 
	> [1:]
	----^`,
			"{}x": `ERROR (code 301): Invalid character "x" non-space byte after top-level value
	in line 1 on file 
	> {}x
	----^`,
			`{"key"}`: `ERROR (code 301): Invalid character "}" after object key
	in line 1 on file 
	> {"key"}
	--------^`,
			`{"key":1:}`: `ERROR (code 301): Invalid character ":" after object key:value pair
	in line 1 on file 
	> {"key":1:}
	----------^`,
			`{"key": 1,:`: `ERROR (code 301): Invalid character ":" looking for beginning of string
	in line 1 on file 
	> {"key": 1,:
	------------^`,
			"{": `ERROR (code 303): Unexpected end of file
	in line 1 on file 
	> {
	--^`,
			"[": `ERROR (code 303): Unexpected end of file
	in line 1 on file 
	> [
	--^`,
			`"string without closing quotation mark`: `ERROR (code 303): Unexpected end of file
	in line 1 on file 
	> "string without closing quotation mark
	---------------------------------------^`,
			`string without opening quotation mark"`: `ERROR (code 301): Invalid character "s" looking for beginning of value
	in line 1 on file 
	> string without opening quotation mark"
	--^`,
			`"str" // comment`: `ERROR (code 301): Invalid character "/" non-space byte after top-level value
	in line 1 on file 
	> "str" // comment
	--------^`,
			"123\n2": `ERROR (code 301): Invalid character "2" non-space byte after top-level value
	in line 2 on file 
	> 2
	--^`,
			"{}-": `ERROR (code 301): Invalid character "-" non-space byte after top-level value
	in line 1 on file 
	> {}-
	----^`,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				d := FromFile(fs.NewFile("", given))

				var foundError bool
				for {
					_, err := d.NextLexeme()
					if err != nil {
						assert.EqualError(t, err, expected)
						foundError = true
						break
					}
				}
				assert.True(t, foundError, "Expects an error")
			})
		}
	})
}
