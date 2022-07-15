package lexeme

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexEventType_IsOpening(t *testing.T) {
	cc := map[LexEventType]bool{
		LiteralBegin:                 true,
		LiteralEnd:                   false,
		ObjectBegin:                  true,
		ObjectEnd:                    false,
		ObjectKeyBegin:               true,
		ObjectKeyEnd:                 false,
		ObjectValueBegin:             true,
		ObjectValueEnd:               false,
		ArrayBegin:                   true,
		ArrayEnd:                     false,
		ArrayItemBegin:               true,
		ArrayItemEnd:                 false,
		InlineAnnotationBegin:        true,
		InlineAnnotationEnd:          false,
		InlineAnnotationTextBegin:    true,
		InlineAnnotationTextEnd:      false,
		MultiLineAnnotationBegin:     true,
		MultiLineAnnotationEnd:       false,
		MultiLineAnnotationTextBegin: true,
		MultiLineAnnotationTextEnd:   false,
		NewLine:                      false,
		TypesShortcutBegin:           true,
		TypesShortcutEnd:             false,
		KeyShortcutBegin:             true,
		KeyShortcutEnd:               false,
		MixedValueBegin:              true,
		MixedValueEnd:                false,
		EndTop:                       false,
	}

	for lexType, expected := range cc {
		t.Run(lexType.String(), func(t *testing.T) {
			assert.Equal(t, expected, lexType.IsOpening())
		})
	}
}

func TestLexEventType_String(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[LexEventType]string{
			LiteralBegin:                 "literal-begin",
			LiteralEnd:                   "literal-end",
			ObjectBegin:                  "object-begin",
			ObjectEnd:                    "object-end",
			ObjectKeyBegin:               "key-begin",
			ObjectKeyEnd:                 "key-end",
			ObjectValueBegin:             "value-begin",
			ObjectValueEnd:               "value-end",
			ArrayBegin:                   "array-begin",
			ArrayEnd:                     "array-end",
			ArrayItemBegin:               "item-begin",
			ArrayItemEnd:                 "item-end",
			InlineAnnotationBegin:        "inline-annotation-begin",
			InlineAnnotationEnd:          "inline-annotation-end",
			InlineAnnotationTextBegin:    "inline-annotation-text-begin",
			InlineAnnotationTextEnd:      "inline-annotation-text-end",
			MultiLineAnnotationBegin:     "multi-line-annotation-begin",
			MultiLineAnnotationEnd:       "multi-line-annotation-end",
			MultiLineAnnotationTextBegin: "multi-line-annotation-text-begin",
			MultiLineAnnotationTextEnd:   "multi-line-annotation-text-end",
			NewLine:                      "new-line",
			TypesShortcutBegin:           "types-shortcut-begin",
			TypesShortcutEnd:             "types-shortcut-end",
			KeyShortcutBegin:             "key-shortcut-begin",
			KeyShortcutEnd:               "key-shortcut-end",
			MixedValueBegin:              "mixed-value-begin",
			MixedValueEnd:                "mixed-value-end",
			EndTop:                       "end-top",
		}

		for lexType, expected := range cc {
			t.Run(expected, func(t *testing.T) {
				assert.Equal(t, expected, lexType.String())
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithValue(t, "Unknown lexical event type", func() {
			_ = LexEventType(255).String()
		})
	})
}
