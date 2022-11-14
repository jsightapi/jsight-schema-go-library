package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/lexeme"

	schema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/internal/mocks"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/ischema/constraint"
	"github.com/jsightapi/jsight-schema-go-library/rules/enum"
)

func Test_newEnumValueLoader(t *testing.T) {
	c := &constraint.Enum{}
	rr := map[string]schema.Rule{
		"foo": mocks.NewRule(t),
	}
	l := newEnumValueLoader(c, rr)

	assert.Same(t, c, l.enumConstraint)
	assert.Equal(t, rr, l.rules)
	assert.NotNil(t, l.stateFunc)
	assert.True(t, l.inProgress)
}

func TestEnumValueLoader_load(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		expected := newFakeLexEvent(lexeme.LiteralBegin)

		l := &enumValueLoader{
			stateFunc: func(lex lexeme.LexEvent) {
				assert.Equal(t, expected, lex)
			},
			inProgress: true,
		}

		ret := l.Load(expected)
		assert.True(t, ret)
	})

	t.Run("negative", func(t *testing.T) {
		assert.Panics(t, func() {
			expected := newFakeLexEvent(lexeme.LiteralBegin)

			l := &enumValueLoader{
				stateFunc: func(lex lexeme.LexEvent) {
					panic("foo")
				},
				inProgress: true,
			}

			ret := l.Load(expected)
			assert.False(t, ret)
		})
	})
}

func TestEnumValueLoader_begin(t *testing.T) {
	cc := map[lexeme.LexEventType]bool{
		lexeme.LiteralBegin:                 true,
		lexeme.LiteralEnd:                   true,
		lexeme.ObjectBegin:                  true,
		lexeme.ObjectEnd:                    true,
		lexeme.ObjectKeyBegin:               true,
		lexeme.ObjectKeyEnd:                 true,
		lexeme.ObjectValueBegin:             true,
		lexeme.ObjectValueEnd:               true,
		lexeme.ArrayBegin:                   false,
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
		lexeme.NewLine:                      true,
		lexeme.TypesShortcutBegin:           true,
		lexeme.TypesShortcutEnd:             true,
		lexeme.KeyShortcutBegin:             true,
		lexeme.KeyShortcutEnd:               true,
		lexeme.MixedValueBegin:              false,
		lexeme.MixedValueEnd:                true,
		lexeme.EndTop:                       true,
	}

	for typ, shouldPanic := range cc {
		t.Run(typ.String(), func(t *testing.T) {
			l := &enumValueLoader{}

			if shouldPanic {
				assert.PanicsWithValue(t, errors.ErrInvalidValueInEnumRule, func() {
					l.begin(newFakeLexEvent(typ))
				})
			} else {
				l.begin(newFakeLexEvent(typ))
				assert.NotNil(t, l.stateFunc)
			}
		})
	}
}

func TestEnumValueLoader_arrayItemBeginOrArrayEnd(t *testing.T) {
	cc := map[lexeme.LexEventType]struct {
		shouldPanic bool
		inProgress  bool
	}{
		lexeme.LiteralBegin: {
			shouldPanic: true,
		},
		lexeme.LiteralEnd: {
			shouldPanic: true,
		},
		lexeme.ObjectBegin: {
			shouldPanic: true,
		},
		lexeme.ObjectEnd: {
			shouldPanic: true,
		},
		lexeme.ObjectKeyBegin: {
			shouldPanic: true,
		},
		lexeme.ObjectKeyEnd: {
			shouldPanic: true,
		},
		lexeme.ObjectValueBegin: {
			shouldPanic: true,
		},
		lexeme.ObjectValueEnd: {
			shouldPanic: true,
		},
		lexeme.ArrayBegin: {
			shouldPanic: true,
		},
		lexeme.ArrayEnd: {
			shouldPanic: false,
			inProgress:  false,
		},
		lexeme.ArrayItemBegin: {
			shouldPanic: false,
			inProgress:  true,
		},
		lexeme.ArrayItemEnd: {
			shouldPanic: true,
		},
		lexeme.InlineAnnotationBegin: {
			shouldPanic: false,
			inProgress:  true,
		},
		lexeme.InlineAnnotationEnd: {
			shouldPanic: true,
		},
		lexeme.InlineAnnotationTextBegin: {
			shouldPanic: true,
		},
		lexeme.InlineAnnotationTextEnd: {
			shouldPanic: true,
		},
		lexeme.MultiLineAnnotationBegin: {
			shouldPanic: true,
		},
		lexeme.MultiLineAnnotationEnd: {
			shouldPanic: true,
		},
		lexeme.MultiLineAnnotationTextBegin: {
			shouldPanic: true,
		},
		lexeme.MultiLineAnnotationTextEnd: {
			shouldPanic: true,
		},
		lexeme.NewLine: {
			shouldPanic: true,
		},
		lexeme.TypesShortcutBegin: {
			shouldPanic: true,
		},
		lexeme.TypesShortcutEnd: {
			shouldPanic: true,
		},
		lexeme.KeyShortcutBegin: {
			shouldPanic: true,
		},
		lexeme.KeyShortcutEnd: {
			shouldPanic: true,
		},
		lexeme.MixedValueBegin: {
			shouldPanic: true,
		},
		lexeme.MixedValueEnd: {
			shouldPanic: true,
		},
		lexeme.EndTop: {
			shouldPanic: true,
		},
	}

	for typ, c := range cc {
		t.Run(typ.String(), func(t *testing.T) {
			l := &enumValueLoader{
				inProgress: true,
			}

			if c.shouldPanic {
				assert.PanicsWithValue(t, errors.ErrLoader, func() {
					l.arrayItemBeginOrArrayEnd(newFakeLexEvent(typ))
				})
			} else {
				l.arrayItemBeginOrArrayEnd(newFakeLexEvent(typ))
				assert.NotNil(t, l.stateFunc)
				assert.Equal(t, c.inProgress, l.inProgress)
			}
		})
	}
}

func TestEnumValueLoader_commentStart(t *testing.T) {
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
		lexeme.InlineAnnotationTextBegin:    false,
		lexeme.InlineAnnotationTextEnd:      true,
		lexeme.MultiLineAnnotationBegin:     true,
		lexeme.MultiLineAnnotationEnd:       true,
		lexeme.MultiLineAnnotationTextBegin: true,
		lexeme.MultiLineAnnotationTextEnd:   true,
		lexeme.NewLine:                      true,
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
			l := &enumValueLoader{}

			if shouldPanic {
				assert.PanicsWithValue(t, errors.ErrLoader, func() {
					l.commentStart(newFakeLexEvent(typ))
				})
			} else {
				l.commentStart(newFakeLexEvent(typ))
				assert.NotNil(t, l.stateFunc)
			}
		})
	}
}

func TestEnumValueLoader_commentEnd(t *testing.T) {
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
		lexeme.InlineAnnotationTextEnd:      false,
		lexeme.MultiLineAnnotationBegin:     true,
		lexeme.MultiLineAnnotationEnd:       true,
		lexeme.MultiLineAnnotationTextBegin: true,
		lexeme.MultiLineAnnotationTextEnd:   true,
		lexeme.NewLine:                      true,
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
			c := constraint.NewEnum()
			c.Append(constraint.NewEnumItem(bytes.NewBytes("42"), ""))

			l := newEnumValueLoader(c, nil)

			if shouldPanic {
				assert.PanicsWithValue(t, errors.ErrLoader, func() {
					l.commentEnd(newFakeLexEvent(typ))
				})
			} else {
				l.commentEnd(newFakeLexEventWithValue(typ, "comment"))

				assert.NotNil(t, l.stateFunc)
				assert.Equal(t, schema.RuleASTNode{
					TokenType:  schema.TokenTypeArray,
					Properties: &schema.RuleASTNodes{},
					Items: []schema.RuleASTNode{
						{
							TokenType:  schema.TokenTypeNumber,
							Value:      "42",
							Comment:    "comment",
							Properties: &schema.RuleASTNodes{},
							Source:     schema.RuleASTNodeSourceManual,
						},
					},
					Source: schema.RuleASTNodeSourceManual,
				}, c.ASTNode())
			}
		})
	}
}

func TestEnumValueLoader_annotationEnd(t *testing.T) {
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
		lexeme.InlineAnnotationEnd:          false,
		lexeme.InlineAnnotationTextBegin:    true,
		lexeme.InlineAnnotationTextEnd:      true,
		lexeme.MultiLineAnnotationBegin:     true,
		lexeme.MultiLineAnnotationEnd:       true,
		lexeme.MultiLineAnnotationTextBegin: true,
		lexeme.MultiLineAnnotationTextEnd:   true,
		lexeme.NewLine:                      true,
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
			l := &enumValueLoader{}

			if shouldPanic {
				assert.PanicsWithValue(t, errors.ErrLoader, func() {
					l.annotationEnd(newFakeLexEvent(typ))
				})
			} else {
				l.annotationEnd(newFakeLexEvent(typ))
				assert.NotNil(t, l.stateFunc)
			}
		})
	}
}

func TestEnumValueLoader_literal(t *testing.T) {
	shouldPanics := []lexeme.LexEventType{
		lexeme.ObjectBegin,
		lexeme.ObjectEnd,
		lexeme.ObjectKeyBegin,
		lexeme.ObjectKeyEnd,
		lexeme.ObjectValueBegin,
		lexeme.ObjectValueEnd,
		lexeme.ArrayBegin,
		lexeme.ArrayEnd,
		lexeme.ArrayItemBegin,
		lexeme.ArrayItemEnd,
		lexeme.InlineAnnotationBegin,
		lexeme.InlineAnnotationEnd,
		lexeme.InlineAnnotationTextBegin,
		lexeme.InlineAnnotationTextEnd,
		lexeme.MultiLineAnnotationBegin,
		lexeme.MultiLineAnnotationEnd,
		lexeme.MultiLineAnnotationTextBegin,
		lexeme.MultiLineAnnotationTextEnd,
		lexeme.NewLine,
		lexeme.TypesShortcutBegin,
		lexeme.TypesShortcutEnd,
		lexeme.KeyShortcutBegin,
		lexeme.KeyShortcutEnd,
		lexeme.MixedValueBegin,
		lexeme.MixedValueEnd,
		lexeme.EndTop,
	}

	for _, typ := range shouldPanics {
		t.Run(typ.String(), func(t *testing.T) {
			assert.PanicsWithValue(t, errors.ErrIncorrectArrayItemTypeInEnumRule, func() {
				(&enumValueLoader{}).literal(newFakeLexEvent(typ))
			})
		})
	}

	t.Run(lexeme.LiteralBegin.String(), func(t *testing.T) {
		l := &enumValueLoader{}
		l.literal(newFakeLexEvent(lexeme.LiteralBegin))
		assert.Nil(t, l.stateFunc)
	})

	t.Run(lexeme.LiteralEnd.String(), func(t *testing.T) {
		c := constraint.NewEnum()
		l := newEnumValueLoader(c, nil)
		l.lastIdx = -1

		l.literal(newFakeLexEventWithValue(lexeme.LiteralEnd, "42"))
		assert.EqualValues(t, 0, l.lastIdx)
		assert.NotNil(t, l.stateFunc)
		assert.Equal(t, schema.RuleASTNode{
			TokenType:  schema.TokenTypeArray,
			Properties: &schema.RuleASTNodes{},
			Items: []schema.RuleASTNode{
				{
					TokenType:  schema.TokenTypeNumber,
					Value:      "42",
					Properties: &schema.RuleASTNodes{},
					Source:     schema.RuleASTNodeSourceManual,
				},
			},
			Source: schema.RuleASTNodeSourceManual,
		}, c.ASTNode())
	})
}

func TestEnumValueLoader_arrayItemEnd(t *testing.T) {
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
		lexeme.ArrayItemEnd:                 false,
		lexeme.InlineAnnotationBegin:        true,
		lexeme.InlineAnnotationEnd:          true,
		lexeme.InlineAnnotationTextBegin:    true,
		lexeme.InlineAnnotationTextEnd:      true,
		lexeme.MultiLineAnnotationBegin:     true,
		lexeme.MultiLineAnnotationEnd:       true,
		lexeme.MultiLineAnnotationTextBegin: true,
		lexeme.MultiLineAnnotationTextEnd:   true,
		lexeme.NewLine:                      true,
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
			l := &enumValueLoader{}

			if shouldPanic {
				assert.PanicsWithValue(t, errors.ErrLoader, func() {
					l.arrayItemEnd(newFakeLexEvent(typ))
				})
			} else {
				l.arrayItemEnd(newFakeLexEvent(typ))
				assert.NotNil(t, l.stateFunc)
			}
		})
	}
}

func TestEnumValueLoader_ruleNameBegin(t *testing.T) {
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
		lexeme.NewLine:                      true,
		lexeme.TypesShortcutBegin:           false,
		lexeme.TypesShortcutEnd:             true,
		lexeme.KeyShortcutBegin:             true,
		lexeme.KeyShortcutEnd:               true,
		lexeme.MixedValueBegin:              true,
		lexeme.MixedValueEnd:                true,
		lexeme.EndTop:                       true,
	}

	for typ, shouldPanic := range cc {
		t.Run(typ.String(), func(t *testing.T) {
			l := &enumValueLoader{}

			if shouldPanic {
				assert.PanicsWithValue(t, errors.ErrLoader, func() {
					l.ruleNameBegin(newFakeLexEvent(typ))
				})
			} else {
				l.ruleNameBegin(newFakeLexEvent(typ))
				assert.NotNil(t, l.stateFunc)
			}
		})
	}
}

func TestEnumValueLoader_ruleName(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		ec := constraint.NewEnum()
		el := newEnumValueLoader(ec, map[string]schema.Rule{
			"foo": enum.New("foo", `[42, 3.14, "foo", false, true, null]`),
		})
		el.stateFunc = nil
		el.ruleName(newFakeLexEventWithValue(lexeme.TypesShortcutEnd, " \nfoo\t \r"))

		assert.NotNil(t, el.stateFunc)
		assert.False(t, el.inProgress)
		assert.Equal(t, "foo", ec.RuleName())
		assert.Equal(t, `enum: [42, 3.14, "foo", false, true, null]`, ec.String())
	})

	t.Run("negative", func(t *testing.T) {
		types := []lexeme.LexEventType{
			lexeme.LiteralBegin,
			lexeme.LiteralEnd,
			lexeme.ObjectBegin,
			lexeme.ObjectEnd,
			lexeme.ObjectKeyBegin,
			lexeme.ObjectKeyEnd,
			lexeme.ObjectValueBegin,
			lexeme.ObjectValueEnd,
			lexeme.ArrayBegin,
			lexeme.ArrayEnd,
			lexeme.ArrayItemBegin,
			lexeme.ArrayItemEnd,
			lexeme.InlineAnnotationBegin,
			lexeme.InlineAnnotationEnd,
			lexeme.InlineAnnotationTextBegin,
			lexeme.InlineAnnotationTextEnd,
			lexeme.MultiLineAnnotationBegin,
			lexeme.MultiLineAnnotationEnd,
			lexeme.MultiLineAnnotationTextBegin,
			lexeme.MultiLineAnnotationTextEnd,
			lexeme.NewLine,
			lexeme.TypesShortcutBegin,
			lexeme.KeyShortcutBegin,
			lexeme.KeyShortcutEnd,
			lexeme.MixedValueBegin,
			lexeme.MixedValueEnd,
			lexeme.EndTop,
		}

		for _, typ := range types {
			t.Run(typ.String(), func(t *testing.T) {
				assert.PanicsWithValue(t, errors.ErrLoader, func() {
					(&enumValueLoader{}).ruleName(newFakeLexEvent(typ))
				})
			})
		}

		cc := map[string]map[string]schema.Rule{
			`Enum rule "ruleName" not found`: {},
			`Rule "ruleName" not an Enum`: {
				"ruleName": mocks.NewRule(t),
			},
			`Invalid enum "ruleName": An array was expected as a value for the "enum"`: {
				"ruleName": enum.New("ruleName", "invalid"),
			},
		}

		for expected, rr := range cc {
			t.Run(expected, func(t *testing.T) {
				assert.PanicsWithError(t, expected, func() {
					(&enumValueLoader{rules: rr}).ruleName(newFakeLexEventWithValue(
						lexeme.TypesShortcutEnd,
						"ruleName",
					))
				})
			})
		}
	})
}

func TestEnumValueLoader_endOfLoading(t *testing.T) {
	types := []lexeme.LexEventType{
		lexeme.LiteralBegin,
		lexeme.LiteralEnd,
		lexeme.ObjectBegin,
		lexeme.ObjectEnd,
		lexeme.ObjectKeyBegin,
		lexeme.ObjectKeyEnd,
		lexeme.ObjectValueBegin,
		lexeme.ObjectValueEnd,
		lexeme.ArrayBegin,
		lexeme.ArrayEnd,
		lexeme.ArrayItemBegin,
		lexeme.ArrayItemEnd,
		lexeme.InlineAnnotationBegin,
		lexeme.InlineAnnotationEnd,
		lexeme.InlineAnnotationTextBegin,
		lexeme.InlineAnnotationTextEnd,
		lexeme.MultiLineAnnotationBegin,
		lexeme.MultiLineAnnotationEnd,
		lexeme.MultiLineAnnotationTextBegin,
		lexeme.MultiLineAnnotationTextEnd,
		lexeme.NewLine,
		lexeme.TypesShortcutBegin,
		lexeme.TypesShortcutEnd,
		lexeme.KeyShortcutBegin,
		lexeme.KeyShortcutEnd,
		lexeme.MixedValueBegin,
		lexeme.MixedValueEnd,
		lexeme.EndTop,
	}

	for _, typ := range types {
		t.Run(typ.String(), func(t *testing.T) {
			assert.PanicsWithValue(t, errors.ErrLoader, func() {
				(&enumValueLoader{}).endOfLoading(newFakeLexEvent(typ))
			})
		})
	}
}

func newFakeLexEvent(t lexeme.LexEventType) lexeme.LexEvent {
	return lexeme.NewLexEvent(t, 0, 0, nil)
}

func newFakeLexEventWithValue(t lexeme.LexEventType, s string) lexeme.LexEvent {
	f := fs.NewFile("", s)
	return lexeme.NewLexEvent(t, 0, bytes.Index(len(s)-1), f)
}
