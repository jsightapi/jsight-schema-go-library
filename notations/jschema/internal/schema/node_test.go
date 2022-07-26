package schema

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/internal/lexeme"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema/constraint"
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

func TestIsOptionalNode(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]struct {
			given    func(*testing.T) Node
			expected bool
		}{
			"node without optional constraints": {
				func(t *testing.T) Node {
					n := NewMockNode(t)
					n.
						On("Constraint", constraint.OptionalConstraintType).
						Return(nil)
					return n
				},
				false,
			},
			"not a bool keeper": {
				func(t *testing.T) Node {
					n := NewMockNode(t)
					n.
						On("Constraint", constraint.OptionalConstraintType).
						Return(constraint.Max{})
					return n
				},
				false,
			},
			"false": {
				func(t *testing.T) Node {
					n := NewMockNode(t)
					n.
						On("Constraint", constraint.OptionalConstraintType).
						Return(constraint.NewOptional([]byte("false")))
					return n
				},
				false,
			},
			"true": {
				func(t *testing.T) Node {
					n := NewMockNode(t)
					n.
						On("Constraint", constraint.OptionalConstraintType).
						Return(constraint.NewOptional([]byte("true")))
					return n
				},
				true,
			},
		}

		for n, c := range cc {
			t.Run(n, func(t *testing.T) {
				actual := IsOptionalNode(c.given(t))
				assert.Equal(t, c.expected, actual)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.Panics(t, func() {
			IsOptionalNode(nil)
		})
	})
}
