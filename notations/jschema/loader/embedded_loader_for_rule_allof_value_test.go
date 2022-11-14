package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/ischema/constraint"
)

func Test_newAllOfValueLoader(t *testing.T) {
	expectedConstraint := &constraint.AllOf{}

	l := newAllOfValueLoader(expectedConstraint)

	assert.Same(t, expectedConstraint, l.allOfConstraint)
	assert.NotNil(t, l.stateFunc)
	assert.True(t, l.inProgress)
}
