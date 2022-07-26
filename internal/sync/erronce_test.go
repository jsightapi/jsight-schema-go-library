package sync

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrOnce_Do(t *testing.T) {
	t.Run("once without error", func(t *testing.T) {
		eo := ErrOnce{}

		called := 0
		fn := func() error {
			called++
			return nil
		}
		t.Cleanup(func() {
			assert.Equal(t, 1, called)
		})

		assert.NoError(t, eo.Do(fn))
		assert.NoError(t, eo.Do(fn))
		assert.NoError(t, eo.Do(fn))
	})

	t.Run("once with error", func(t *testing.T) {
		eo := ErrOnce{}

		called := 0
		fn := func() error {
			called++
			return fmt.Errorf("fake error %d", called)
		}
		t.Cleanup(func() {
			assert.Equal(t, 1, called)
		})

		assert.EqualError(t, eo.Do(fn), "fake error 1")
		assert.EqualError(t, eo.Do(fn), "fake error 1")
		assert.EqualError(t, eo.Do(fn), "fake error 1")
	})
}

func TestErrOnceWithValue_Do(t *testing.T) {
	t.Run("once without error", func(t *testing.T) {
		eo := ErrOnceWithValue[int]{}

		called := 0
		fn := func() (int, error) {
			called++
			return called, nil
		}
		t.Cleanup(func() {
			assert.Equal(t, 1, called)
		})

		v, err := eo.Do(fn)
		assert.Equal(t, 1, v)
		assert.NoError(t, err)

		v, err = eo.Do(fn)
		assert.Equal(t, 1, v)
		assert.NoError(t, err)

		v, err = eo.Do(fn)
		assert.Equal(t, 1, v)
		assert.NoError(t, err)
	})

	t.Run("once with error", func(t *testing.T) {
		eo := ErrOnceWithValue[int]{}

		called := 0
		fn := func() (int, error) {
			called++
			return 0, fmt.Errorf("fake error %d", called)
		}
		t.Cleanup(func() {
			assert.Equal(t, 1, called)
		})

		v, err := eo.Do(fn)
		assert.Equal(t, 0, v)
		assert.EqualError(t, err, "fake error 1")

		v, err = eo.Do(fn)
		assert.Equal(t, 0, v)
		assert.EqualError(t, err, "fake error 1")

		v, err = eo.Do(fn)
		assert.Equal(t, 0, v)
		assert.EqualError(t, err, "fake error 1")
	})
}
