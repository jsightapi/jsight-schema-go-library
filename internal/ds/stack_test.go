package ds

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack_Len(t *testing.T) {
	cc := map[string]struct {
		stack    *Stack[int]
		expected int
	}{
		"nil": {},
		"filled": {
			stack:    &Stack[int]{vals: []int{1, 2, 3}},
			expected: 3,
		},
	}

	for n, c := range cc {
		t.Run(n, func(t *testing.T) {
			actual := c.stack.Len()
			assert.Equal(t, c.expected, actual)
		})
	}
}

func TestStack_Push(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		s := &Stack[int]{}

		s.Push(1)
		s.Push(2)
		s.Push(3)

		assert.Equal(t, []int{1, 2, 3}, s.vals)
	})

	t.Run("negative", func(t *testing.T) {
		assert.Panics(t, func() {
			var s *Stack[int]
			s.Push(0)
		})
	})
}

func TestStack_Pop(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		s := &Stack[int]{vals: []int{1, 2, 3}}

		assert.Equal(t, 3, s.Pop())
		assert.Equal(t, 2, s.Pop())
		assert.Equal(t, 1, s.Pop())

		assert.Equal(t, []int{}, s.vals)
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithValue(t, "Reading from empty stack", func() {
			var s *Stack[int]
			s.Pop()
		})
	})
}

func TestStack_Peek(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		s := &Stack[int]{vals: []int{1, 2, 3}}

		assert.Equal(t, 3, s.Peek())
		assert.Equal(t, 3, s.Peek())
		assert.Equal(t, 3, s.Peek())

		assert.Equal(t, []int{1, 2, 3}, s.vals)
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithValue(t, "Reading from empty stack", func() {
			(&Stack[int]{}).Peek()
		})
	})
}

func TestStack_Get(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		s := &Stack[int]{vals: []int{1, 2, 3}}

		assert.Equal(t, 3, s.Get(2))
		assert.Equal(t, 2, s.Get(1))
		assert.Equal(t, 1, s.Get(0))

		assert.Equal(t, []int{1, 2, 3}, s.vals)
	})

	t.Run("negative", func(t *testing.T) {
		cc := []int{
			-1,
			2,
		}

		for _, i := range cc {
			t.Run(strconv.Itoa(i), func(t *testing.T) {
				assert.PanicsWithValue(t, "Reading a nonexistent element of the stack", func() {
					(&Stack[int]{vals: []int{1}}).Get(i)
				})
			})
		}
	})
}
