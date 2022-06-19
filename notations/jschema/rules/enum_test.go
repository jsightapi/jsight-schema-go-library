package rules

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/internal/lexeme"
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
			"42":             2,
			"42 [] foo":      2,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				actual, err := NewEnum("", []byte(given)).Len()
				require.NoError(t, err)
				assert.Equal(t, expected, actual)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		_, err := NewEnum("", []byte("[")).Len()
		assert.EqualError(t, err, `ERROR (code 303): Unexpected end of file
	in line 1 on file 
	> [
	--^`)
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
				err := NewEnum("enum", []byte(enum)).Check()
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

func TestEnum_Values(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		vals, err := NewEnum("", []byte(`[
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
		_, err := NewEnum("", []byte("123")).Values()
		assert.EqualError(t, err, `ERROR (code 1600): An array was expected as a value for the "enum"
	in line 1 on file 
	> 123
	--^`)
	})
}

func Test_newEnumChecker(t *testing.T) {
	c := newEnumChecker()

	assert.NotNil(t, c.stateFunc)
}

func TestEnumChecker_begin(t *testing.T) {
	type expecter func(t *testing.T, c *enumChecker, lex lexeme.LexEvent)

	expectPanic := func(t *testing.T, c *enumChecker, lex lexeme.LexEvent) {
		assert.PanicsWithValue(t, errors.ErrEnumArrayExpected, func() {
			c.begin(lex)
		})
	}

	cc := map[lexeme.LexEventType]expecter{
		lexeme.LiteralBegin:     expectPanic,
		lexeme.LiteralEnd:       expectPanic,
		lexeme.ObjectBegin:      expectPanic,
		lexeme.ObjectEnd:        expectPanic,
		lexeme.ObjectKeyBegin:   expectPanic,
		lexeme.ObjectKeyEnd:     expectPanic,
		lexeme.ObjectValueBegin: expectPanic,
		lexeme.ObjectValueEnd:   expectPanic,
		lexeme.ArrayBegin: func(t *testing.T, c *enumChecker, lex lexeme.LexEvent) {
			c.begin(lex)
			assert.NotNil(t, c.stateFunc)
		},
		lexeme.ArrayEnd:                     expectPanic,
		lexeme.ArrayItemBegin:               expectPanic,
		lexeme.ArrayItemEnd:                 expectPanic,
		lexeme.InlineAnnotationBegin:        expectPanic,
		lexeme.InlineAnnotationEnd:          expectPanic,
		lexeme.InlineAnnotationTextBegin:    expectPanic,
		lexeme.InlineAnnotationTextEnd:      expectPanic,
		lexeme.MultiLineAnnotationBegin:     expectPanic,
		lexeme.MultiLineAnnotationEnd:       expectPanic,
		lexeme.MultiLineAnnotationTextBegin: expectPanic,
		lexeme.MultiLineAnnotationTextEnd:   expectPanic,
		lexeme.NewLine: func(t *testing.T, c *enumChecker, lex lexeme.LexEvent) {
			c.begin(lex)
		},
		lexeme.TypesShortcutBegin: expectPanic,
		lexeme.TypesShortcutEnd:   expectPanic,
		lexeme.KeyShortcutBegin:   expectPanic,
		lexeme.KeyShortcutEnd:     expectPanic,
		lexeme.MixedValueBegin:    expectPanic,
		lexeme.MixedValueEnd:      expectPanic,
		lexeme.EndTop:             expectPanic,
	}

	for typ, fn := range cc {
		t.Run(typ.String(), func(t *testing.T) {
			fn(t, &enumChecker{}, newLexeme(typ))
		})
	}
}

func TestEnumChecker_arrayItemBeginOrArrayEnd(t *testing.T) {
	type expecter func(t *testing.T, c *enumChecker, lex lexeme.LexEvent)

	expectPanic := func(t *testing.T, c *enumChecker, lex lexeme.LexEvent) {
		assert.PanicsWithValue(t, errors.ErrLoader, func() {
			c.arrayItemBeginOrArrayEnd(lex)
		})
	}

	cc := map[lexeme.LexEventType]expecter{
		lexeme.LiteralBegin:     expectPanic,
		lexeme.LiteralEnd:       expectPanic,
		lexeme.ObjectBegin:      expectPanic,
		lexeme.ObjectEnd:        expectPanic,
		lexeme.ObjectKeyBegin:   expectPanic,
		lexeme.ObjectKeyEnd:     expectPanic,
		lexeme.ObjectValueBegin: expectPanic,
		lexeme.ObjectValueEnd:   expectPanic,
		lexeme.ArrayBegin:       expectPanic,
		lexeme.ArrayEnd: func(t *testing.T, c *enumChecker, lex lexeme.LexEvent) {
			c.arrayItemBeginOrArrayEnd(lex)
			assert.NotNil(t, c.stateFunc)
		},
		lexeme.ArrayItemBegin: func(t *testing.T, c *enumChecker, lex lexeme.LexEvent) {
			c.arrayItemBeginOrArrayEnd(lex)
			assert.NotNil(t, c.stateFunc)
		},
		lexeme.ArrayItemEnd:                 expectPanic,
		lexeme.InlineAnnotationBegin:        expectPanic,
		lexeme.InlineAnnotationEnd:          expectPanic,
		lexeme.InlineAnnotationTextBegin:    expectPanic,
		lexeme.InlineAnnotationTextEnd:      expectPanic,
		lexeme.MultiLineAnnotationBegin:     expectPanic,
		lexeme.MultiLineAnnotationEnd:       expectPanic,
		lexeme.MultiLineAnnotationTextBegin: expectPanic,
		lexeme.MultiLineAnnotationTextEnd:   expectPanic,
		lexeme.NewLine: func(t *testing.T, c *enumChecker, lex lexeme.LexEvent) {
			c.arrayItemBeginOrArrayEnd(lex)
		},
		lexeme.TypesShortcutBegin: expectPanic,
		lexeme.TypesShortcutEnd:   expectPanic,
		lexeme.KeyShortcutBegin:   expectPanic,
		lexeme.KeyShortcutEnd:     expectPanic,
		lexeme.MixedValueBegin:    expectPanic,
		lexeme.MixedValueEnd:      expectPanic,
		lexeme.EndTop:             expectPanic,
	}

	for typ, fn := range cc {
		t.Run(typ.String(), func(t *testing.T) {
			fn(t, &enumChecker{}, newLexeme(typ))
		})
	}
}

func TestEnumChecker_literal(t *testing.T) {
	type expecter func(t *testing.T, c *enumChecker, lex lexeme.LexEvent)

	expectPanic := func(t *testing.T, c *enumChecker, lex lexeme.LexEvent) {
		assert.PanicsWithValue(t, errors.ErrIncorrectArrayItemTypeInEnumRule, func() {
			c.literal(lex)
		})
	}

	cc := map[lexeme.LexEventType]expecter{
		lexeme.LiteralBegin: func(t *testing.T, c *enumChecker, lex lexeme.LexEvent) {
			c.literal(lex)
		},
		lexeme.LiteralEnd: func(t *testing.T, c *enumChecker, lex lexeme.LexEvent) {
			c.literal(lex)
			assert.NotNil(t, c.stateFunc)
		},
		lexeme.ObjectBegin:                  expectPanic,
		lexeme.ObjectEnd:                    expectPanic,
		lexeme.ObjectKeyBegin:               expectPanic,
		lexeme.ObjectKeyEnd:                 expectPanic,
		lexeme.ObjectValueBegin:             expectPanic,
		lexeme.ObjectValueEnd:               expectPanic,
		lexeme.ArrayBegin:                   expectPanic,
		lexeme.ArrayEnd:                     expectPanic,
		lexeme.ArrayItemBegin:               expectPanic,
		lexeme.ArrayItemEnd:                 expectPanic,
		lexeme.InlineAnnotationBegin:        expectPanic,
		lexeme.InlineAnnotationEnd:          expectPanic,
		lexeme.InlineAnnotationTextBegin:    expectPanic,
		lexeme.InlineAnnotationTextEnd:      expectPanic,
		lexeme.MultiLineAnnotationBegin:     expectPanic,
		lexeme.MultiLineAnnotationEnd:       expectPanic,
		lexeme.MultiLineAnnotationTextBegin: expectPanic,
		lexeme.MultiLineAnnotationTextEnd:   expectPanic,
		lexeme.NewLine:                      expectPanic,
		lexeme.TypesShortcutBegin:           expectPanic,
		lexeme.TypesShortcutEnd:             expectPanic,
		lexeme.KeyShortcutBegin:             expectPanic,
		lexeme.KeyShortcutEnd:               expectPanic,
		lexeme.MixedValueBegin:              expectPanic,
		lexeme.MixedValueEnd:                expectPanic,
		lexeme.EndTop:                       expectPanic,
	}

	for typ, fn := range cc {
		t.Run(typ.String(), func(t *testing.T) {
			fn(t, &enumChecker{}, newLexeme(typ))
		})
	}
}

func TestEnumChecker_arrayItemEnd(t *testing.T) {
	type expecter func(t *testing.T, c *enumChecker, lex lexeme.LexEvent)

	expectPanic := func(t *testing.T, c *enumChecker, lex lexeme.LexEvent) {
		assert.PanicsWithValue(t, errors.ErrLoader, func() {
			c.arrayItemEnd(lex)
		})
	}

	cc := map[lexeme.LexEventType]expecter{
		lexeme.LiteralBegin:     expectPanic,
		lexeme.LiteralEnd:       expectPanic,
		lexeme.ObjectBegin:      expectPanic,
		lexeme.ObjectEnd:        expectPanic,
		lexeme.ObjectKeyBegin:   expectPanic,
		lexeme.ObjectKeyEnd:     expectPanic,
		lexeme.ObjectValueBegin: expectPanic,
		lexeme.ObjectValueEnd:   expectPanic,
		lexeme.ArrayBegin:       expectPanic,
		lexeme.ArrayEnd:         expectPanic,
		lexeme.ArrayItemBegin:   expectPanic,
		lexeme.ArrayItemEnd: func(t *testing.T, c *enumChecker, lex lexeme.LexEvent) {
			c.arrayItemEnd(lex)
			assert.NotNil(t, c.stateFunc)
		},
		lexeme.InlineAnnotationBegin:        expectPanic,
		lexeme.InlineAnnotationEnd:          expectPanic,
		lexeme.InlineAnnotationTextBegin:    expectPanic,
		lexeme.InlineAnnotationTextEnd:      expectPanic,
		lexeme.MultiLineAnnotationBegin:     expectPanic,
		lexeme.MultiLineAnnotationEnd:       expectPanic,
		lexeme.MultiLineAnnotationTextBegin: expectPanic,
		lexeme.MultiLineAnnotationTextEnd:   expectPanic,
		lexeme.NewLine: func(t *testing.T, c *enumChecker, lex lexeme.LexEvent) {
			c.arrayItemEnd(lex)
		},
		lexeme.TypesShortcutBegin: expectPanic,
		lexeme.TypesShortcutEnd:   expectPanic,
		lexeme.KeyShortcutBegin:   expectPanic,
		lexeme.KeyShortcutEnd:     expectPanic,
		lexeme.MixedValueBegin:    expectPanic,
		lexeme.MixedValueEnd:      expectPanic,
		lexeme.EndTop:             expectPanic,
	}

	for typ, fn := range cc {
		t.Run(typ.String(), func(t *testing.T) {
			fn(t, &enumChecker{}, newLexeme(typ))
		})
	}
}

func TestEnumChecker_afterEndOfEnum(t *testing.T) {
	cc := map[lexeme.LexEventType]bool{
		lexeme.LiteralBegin:                 true,
		lexeme.LiteralEnd:                   true,
		lexeme.ObjectBegin:                  true,
		lexeme.ObjectEnd:                    true,
		lexeme.ObjectKeyBegin:               true,
		lexeme.ObjectKeyEnd:                 true,
		lexeme.ObjectValueBegin:             true,
		lexeme.ObjectValueEnd:               true,
		lexeme.ArrayBegin:                   true,
		lexeme.ArrayEnd:                     true,
		lexeme.ArrayItemBegin:               true,
		lexeme.ArrayItemEnd:                 true,
		lexeme.InlineAnnotationBegin:        true,
		lexeme.InlineAnnotationEnd:          true,
		lexeme.InlineAnnotationTextBegin:    true,
		lexeme.InlineAnnotationTextEnd:      true,
		lexeme.MultiLineAnnotationBegin:     true,
		lexeme.MultiLineAnnotationEnd:       true,
		lexeme.MultiLineAnnotationTextBegin: true,
		lexeme.MultiLineAnnotationTextEnd:   true,
		lexeme.NewLine:                      false,
		lexeme.TypesShortcutBegin:           true,
		lexeme.TypesShortcutEnd:             true,
		lexeme.KeyShortcutBegin:             true,
		lexeme.KeyShortcutEnd:               true,
		lexeme.MixedValueBegin:              true,
		lexeme.MixedValueEnd:                true,
		lexeme.EndTop:                       true,
	}

	for typ, shouldPanic := range cc {
		t.Run(typ.String(), func(t *testing.T) {
			c := &enumChecker{}
			lex := newLexeme(typ)

			if shouldPanic {
				assert.PanicsWithValue(t, errors.ErrUnnecessaryLexemeAfterTheEndOfEnum, func() {
					c.afterEndOfEnum(lex)
				})
			} else {
				c.afterEndOfEnum(lex)
			}
		})
	}
}

func newLexeme(t lexeme.LexEventType) lexeme.LexEvent {
	return lexeme.NewLexEvent(t, 0, 1, fs.NewFile("foo", []byte("123456")))
}
