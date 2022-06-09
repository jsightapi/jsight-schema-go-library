package kit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/errors"
)

func requireNoError(t *testing.T, err Error) {
	if assert.Nil(t, err) {
		return
	}

	t.Errorf("Got unexpected error (%d) %s", err.ErrCode(), err.Message())
}

func assertKitError(
	t *testing.T,
	err Error,
	expectedFilename string,
	expectedPosition uint,
	expectedMessage string,
	expectedCode errors.ErrorCode,
) {
	if !assert.NotNil(t, err) {
		return
	}

	assert.Equal(t, expectedFilename, err.Filename())
	assert.Equal(t, expectedPosition, err.Position())
	assert.Equal(t, expectedMessage, err.Message())
	assert.Equal(t, int(expectedCode), err.ErrCode())
}
