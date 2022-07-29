package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema/constraint"
)

func Test_newAllOfValueLoader(t *testing.T) {
	expectedConstraint := &constraint.AllOf{}

	el := newAllOfValueLoader(expectedConstraint)

	require.IsType(t, &allOfValueLoader{}, el)

	l := el.(*allOfValueLoader)
	assert.Same(t, expectedConstraint, l.allOfConstraint)
	assert.NotNil(t, l.stateFunc)
	assert.True(t, l.inProgress)
}
