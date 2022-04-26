package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
	"github.com/jsightapi/jsight-schema-go-library/internal/lexeme"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema/constraint"
)

func TestNewMixedValueNode(t *testing.T) {
	e := lexeme.NewLexEvent(lexeme.MixedValueBegin, 0, 0, nil)
	n := NewMixedValueNode(e)

	assert.Nil(t, n.parent)
	assert.Equal(t, json.TypeMixed, n.jsonType)
	assert.Equal(t, e, n.schemaLexEvent)
	assert.Equal(t, &Constraints{}, n.constraints)
}

func TestMixedValueNode_Grow(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		n := NewMixedValueNode(lexeme.NewLexEvent(lexeme.MixedValueBegin, 0, 0, nil))
		n.parent = newObjectNode(lexeme.NewLexEvent(lexeme.ObjectBegin, 0, 0, nil))

		cc := map[lexeme.LexEventType]Node{
			lexeme.MixedValueBegin: n,
			lexeme.MixedValueEnd:   n.parent,
		}

		for lexType, expected := range cc {
			t.Run(lexType.String(), func(t *testing.T) {
				actual, ok := n.Grow(lexeme.NewLexEvent(lexType, 0, 0, fs.NewFile("", []byte("foo"))))
				assert.Equal(t, expected, actual)
				assert.False(t, ok)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithValue(t,
			`Unexpected lexical event "`+lexeme.ObjectBegin.String()+`" in mixed value node`,
			func() {
				NewMixedValueNode(lexeme.NewLexEvent(lexeme.MixedValueBegin, 0, 0, nil)).
					Grow(lexeme.NewLexEvent(lexeme.ObjectBegin, 0, 0, nil))
			},
		)
	})
}

func TestMixedValueNode_IndentedTreeString(t *testing.T) {
	TestMixedValueNode_IndentedNodeString(t)
}

func TestMixedValueNode_IndentedNodeString(t *testing.T) {
	createNode := func(
		constraints map[constraint.Type]constraint.Constraint,
		order []constraint.Type,
	) *MixedValueNode {
		n := NewMixedValueNode(lexeme.NewLexEvent(lexeme.MixedValueBegin, 0, 0, nil))
		if constraints != nil {
			n.constraints = &Constraints{
				data:  constraints,
				order: order,
			}
		}
		return n
	}

	cc := map[string]struct {
		n        *MixedValueNode
		depth    int
		expected string
	}{
		"depth 0, without constraints": {
			createNode(nil, nil),
			0,
			"* mixed\n",
		},

		"depth 1, without constraints": {
			createNode(nil, nil),
			1,
			"\t* mixed\n",
		},

		"depth 0, with constraints": {
			createNode(map[constraint.Type]constraint.Constraint{
				constraint.ConstType: &constraint.Const{},
			}, []constraint.Type{constraint.ConstType}),
			0,
			"* mixed\n* const: false\n",
		},

		"depth 1, with constraints": {
			createNode(map[constraint.Type]constraint.Constraint{
				constraint.ConstType: &constraint.Const{},
			}, []constraint.Type{constraint.ConstType}),
			1,
			"\t* mixed\n\t* const: false\n",
		},
	}

	for name, c := range cc {
		t.Run(name, func(t *testing.T) {
			actual := c.n.IndentedNodeString(c.depth)
			assert.Equal(t, c.expected, actual)
		})
	}
}
