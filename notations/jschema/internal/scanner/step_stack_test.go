package scanner

import (
	"testing"
)

var foo stepFunc = func(scanner *Scanner, b byte) state {
	return scanContinue
}

func TestStepStackLen(t *testing.T) {
	stack := make(stepFuncStack, 0, 3)
	if stack.Len() != 0 {
		t.Error("Incorrect stack length")
	}

	stack.Push(foo)
	stack.Push(foo)
	if stack.Len() != 2 {
		t.Error("Incorrect stack length")
	}

	stack.Peek()
	if stack.Len() != 2 {
		t.Error("Incorrect stack length")
	}

	stack.Get(0)
	if stack.Len() != 2 {
		t.Error("Incorrect stack length")
	}

	stack.Pop()
	if stack.Len() != 1 {
		t.Error("Incorrect stack length")
	}
}

func TestStepStackPush(t *testing.T) {
	stack := make(stepFuncStack, 0, 2)
	stack.Push(foo)
	stack.Push(foo)
	stack.Push(foo) // reallocate slice
	if stack.Len() != 3 {
		t.Error("Incorrect stack length")
	}
}

func TestStepStackPeekError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expect error")
		}
	}()
	stack := make(stepFuncStack, 0)
	stack.Peek()
}

func TestStepStackGetError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expect error")
		}
	}()
	stack := make(stepFuncStack, 0)
	stack.Get(1)
}
