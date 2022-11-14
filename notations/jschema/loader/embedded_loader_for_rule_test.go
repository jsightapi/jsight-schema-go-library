package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"

	schema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/internal/mocks"
	jschemaMocks "github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/mocks"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/ischema"
)

func Test_newRuleLoader(t *testing.T) {
	n := jschemaMocks.NewNode(t)
	s := &ischema.ISchema{}
	rules := map[string]schema.Rule{
		"foo": mocks.NewRule(t),
	}

	l := newRuleLoader(n, 1, s, rules)

	assert.Same(t, n, l.node)
	assert.Same(t, s, l.rootSchema)
	assert.Equal(t, rules, l.rules)
	assert.EqualValues(t, 1, l.nodesPerCurrentLineCount)
	assert.NotNil(t, l.stateFunc)
}
