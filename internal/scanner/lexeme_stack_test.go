package scanner

import (
	"j/schema/fs"
	"j/schema/internal/lexeme"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexemeEventStack_Len(t *testing.T) {
	lex := lexeme.NewLexEvent(lexeme.NewLine, 0, 0, &fs.File{})

	stack := make(LexemesStack, 0, 3)
	assert.Equal(t, 0, stack.Len())

	stack.Push(lex)
	stack.Push(lex)
	assert.Equal(t, 2, stack.Len())

	stack.Peek()
	assert.Equal(t, 2, stack.Len())

	stack.Get(0)
	assert.Equal(t, 2, stack.Len())

	stack.Pop()
	assert.Equal(t, 1, stack.Len())
}

func TestLexemeEventStack_Peek(t *testing.T) {
	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithValue(t, "Reading from empty stack", func() {
			(&LexemesStack{}).Peek()
		})
	})
}

func TestLexemeEventStack_Get(t *testing.T) {
	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithValue(t, "Reading a nonexistent element of the stack", func() {
			(&LexemesStack{}).Get(1)
		})
	})
}
