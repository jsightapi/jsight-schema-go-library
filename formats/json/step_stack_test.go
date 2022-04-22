package json

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var foo stepFunc = func(scanner *scanner, b byte) state {
	return scanContinue
}

func TestStepFuncStack_Len(t *testing.T) {
	stack := make(stepFuncStack, 0, 3)
	assert.Equal(t, 0, stack.Len())

	stack.Push(foo)
	stack.Push(foo)
	assert.Equal(t, 2, stack.Len())

	stack.Peek()
	assert.Equal(t, 2, stack.Len())

	stack.Get(0)
	assert.Equal(t, 2, stack.Len())

	stack.Pop()
	assert.Equal(t, 1, stack.Len())
}

func TestStepFuncStack_Push(t *testing.T) {
	stack := make(stepFuncStack, 0, 2)
	stack.Push(foo)
	stack.Push(foo)
	stack.Push(foo) // reallocate slice
	assert.Equal(t, 3, stack.Len())
}

func TestStepFuncStack_Peek(t *testing.T) {
	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithValue(t, "Reading from empty stack", func() {
			(&stepFuncStack{}).Peek()
		})
	})
}

func TestStepStackGetError(t *testing.T) {
	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithValue(t, "Reading a nonexistent element of the stack", func() {
			(&stepFuncStack{}).Get(1)
		})
	})
}
