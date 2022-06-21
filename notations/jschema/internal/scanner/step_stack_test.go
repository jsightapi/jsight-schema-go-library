package scanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var foo stepFunc = func(scanner *Scanner, b byte) state {
	return scanContinue
}

func TestStepStack_Push(t *testing.T) {
	stack := make(stepFuncStack, 0, 2)
	stack.Push(foo)
	stack.Push(foo)
	stack.Push(foo) // reallocate slice
	assert.Len(t, stack, 3)
}

func TestStepStack_Peek(t *testing.T) {
	t.Run("negative", func(t *testing.T) {
		assert.Panics(t, func() {
			stack := make(stepFuncStack, 0)
			stack.Peek()
		})
	})
}
