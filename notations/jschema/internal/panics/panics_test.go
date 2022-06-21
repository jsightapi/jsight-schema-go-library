package panics

import (
	stdErrors "errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandle(t *testing.T) {
	cc := map[string]struct {
		r           interface{}
		originErr   error
		expectedErr error
	}{
		"r nil, originErr nil": {
			r:           nil,
			originErr:   nil,
			expectedErr: nil,
		},

		"r nil, originErr isn't nil": {
			r:           nil,
			originErr:   stdErrors.New("origin fake error"),
			expectedErr: stdErrors.New("origin fake error"),
		},

		"r isn't nil, originErr nil": {
			r:           stdErrors.New("r fake error"),
			originErr:   nil,
			expectedErr: stdErrors.New("r fake error"),
		},

		"r isn't nil, originErr isn't nil": {
			r:           stdErrors.New("r fake error"),
			originErr:   stdErrors.New("origin fake error"),
			expectedErr: stdErrors.New("origin fake error"),
		},

		"r isn't nil not error, originErr isn't nil": {
			r:           "foo",
			originErr:   stdErrors.New("origin fake error"),
			expectedErr: stdErrors.New("origin fake error"),
		},
	}

	for n, c := range cc {
		t.Run(n, func(t *testing.T) {
			err := Handle(c.r, c.originErr)
			assert.Equal(t, c.expectedErr, err)
		})
	}

	t.Run("r isn't nil not error, originErr nil", func(t *testing.T) {
		assert.PanicsWithValue(t, "foo", func() {
			_ = Handle("foo", nil)
		})
	})
}
