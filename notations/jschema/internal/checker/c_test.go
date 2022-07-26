package checker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/mocks"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema"
)

func Test_newNodeChecker(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[schema.Node]nodeChecker{
			&schema.LiteralNode{}: literalChecker{},
			&schema.ObjectNode{}:  objectChecker{},
			&schema.ArrayNode{}:   arrayChecker{},
			&schema.MixedNode{}:   mixedChecker{},
		}

		for node, expected := range cc {
			t.Run(fmt.Sprintf("%T", node), func(t *testing.T) {
				actual, err := newNodeChecker(node)
				require.NoError(t, err)

				assert.IsType(t, expected, actual)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]schema.Node{
			"nil":          nil,
			"not expected": &mocks.Node{},
		}

		for n, c := range cc {
			t.Run(n, func(t *testing.T) {
				_, err := newNodeChecker(c)
				assert.ErrorIs(t, err, errors.ErrImpossible)
			})
		}
	})
}
