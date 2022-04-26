package schema

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/internal/lexeme"
)

func TestNewNode(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[lexeme.LexEventType]Node{
			lexeme.LiteralBegin:    &LiteralNode{},
			lexeme.ObjectBegin:     &ObjectNode{},
			lexeme.ArrayBegin:      &ArrayNode{},
			lexeme.MixedValueBegin: &MixedValueNode{},
		}

		for lexType, expected := range cc {
			t.Run(lexType.String(), func(t *testing.T) {
				actual := NewNode(lexeme.NewLexEvent(lexType, 0, 0, nil))
				assert.IsType(t, expected, actual)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		const lexType = lexeme.NewLine
		assert.PanicsWithValue(t, fmt.Sprintf("Can not create node from the lexical event %q", lexType), func() {
			NewNode(lexeme.NewLexEvent(lexeme.NewLine, 0, 0, nil))
		})
	})
}
