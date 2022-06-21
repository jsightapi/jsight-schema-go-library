package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/internal/mocks"
	jschemaMocks "github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/mocks"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema"
)

func Test_newRuleLoader(t *testing.T) {
	n := jschemaMocks.NewNode(t)
	s := &schema.Schema{}
	rules := map[string]jschema.Rule{
		"foo": mocks.NewRule(t),
	}

	l := newRuleLoader(n, 1, s, rules)

	assert.Same(t, n, l.node)
	assert.Same(t, s, l.rootSchema)
	assert.Equal(t, rules, l.rules)
	assert.EqualValues(t, 1, l.nodesPerCurrentLineCount)
	assert.NotNil(t, l.stateFunc)
}
