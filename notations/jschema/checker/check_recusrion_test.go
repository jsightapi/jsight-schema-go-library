package checker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// CheckRecursion was tested in jschema.

func TestRecursionChecker_visit(t *testing.T) {
	rc := &recursionChecker{
		visited: map[string]struct{}{},
	}

	assert.True(t, rc.visit("foo"))

	assert.Len(t, rc.visited, 1)
	assert.Contains(t, rc.visited, "foo")
	assert.Equal(t, []string{"foo"}, rc.path)

	assert.True(t, rc.visit("bar"))

	assert.Len(t, rc.visited, 2)
	assert.Contains(t, rc.visited, "foo")
	assert.Contains(t, rc.visited, "bar")
	assert.Equal(t, []string{"foo", "bar"}, rc.path)

	assert.False(t, rc.visit("foo"))

	assert.Len(t, rc.visited, 2)
	assert.Contains(t, rc.visited, "foo")
	assert.Contains(t, rc.visited, "bar")
	assert.Equal(t, []string{"foo", "bar", "foo"}, rc.path)
}

func TestRecursionChecker_leave(t *testing.T) {
	rc := &recursionChecker{
		visited: map[string]struct{}{
			"foo": {},
			"bar": {},
		},
		path: []string{"foo", "bar"},
	}

	rc.leave("bar")

	assert.Len(t, rc.visited, 1)
	assert.Contains(t, rc.visited, "foo")
	assert.Equal(t, []string{"foo"}, rc.path)

	assert.True(t, rc.visit("bar"))

	assert.Len(t, rc.visited, 2)
	assert.Contains(t, rc.visited, "foo")
	assert.Contains(t, rc.visited, "bar")
	assert.Equal(t, []string{"foo", "bar"}, rc.path)

	assert.False(t, rc.visit("foo"))

	assert.Len(t, rc.visited, 2)
	assert.Contains(t, rc.visited, "foo")
	assert.Contains(t, rc.visited, "bar")
	assert.Equal(t, []string{"foo", "bar", "foo"}, rc.path)
}

func TestRecursionChecker_createError(t *testing.T) {
	err := (&recursionChecker{
		path: []string{"@foo", "@bar", "@fizz", "@buzz", "@foo"},
	}).
		createError()

	assert.EqualError(t, err, "Infinity recursion detected @foo -> @bar -> @fizz -> @buzz -> @foo")
}
