package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jschema "github.com/jsightapi/jsight-schema-go-library"
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

func TestMixedValueNode_AddConstraint(t *testing.T) {
	t.Run("Type constraint", func(t *testing.T) {
		n := createFakeMixedValueNode()
		n.AddConstraint(createFakeTypeConstraint("@foo"))

		assert.Equal(t, []string{"@foo"}, n.types)
	})

	t.Run("Or constraint", func(t *testing.T) {
		n := createFakeMixedValueNode()
		n.AddConstraint(constraint.NewOr(jschema.RuleASTNodeSourceManual))

		assert.Equal(t, []string(nil), n.types)
	})

	t.Run("TypeList constraint", func(t *testing.T) {
		c := constraint.NewTypesList(jschema.RuleASTNodeSourceManual)
		c.AddName("@foo", "@foo", jschema.RuleASTNodeSourceManual)
		c.AddName("@bar", "@bar", jschema.RuleASTNodeSourceManual)

		n := createFakeMixedValueNode()
		n.AddConstraint(c)

		assert.Equal(t, []string{"@foo", "@bar"}, n.types)
	})
}

func TestMixedValueNode_addTypeConstraint(t *testing.T) {
	t.Run("not exists", func(t *testing.T) {
		const name = "@foo"
		n := createFakeMixedValueNode()
		n.addTypeConstraint(createFakeTypeConstraint(name))

		c, ok := n.baseNode.constraints.Get(constraint.TypeConstraintType)
		require.True(t, ok)
		require.IsType(t, &constraint.TypeConstraint{}, c)

		assert.Equal(t, name, c.(*constraint.TypeConstraint).Bytes().String())
	})

	t.Run("exists", func(t *testing.T) {
		cc := map[string]struct {
			exists             *constraint.TypeConstraint
			new                *constraint.TypeConstraint
			expected           *constraint.TypeConstraint
			expectedSchemaType string
		}{
			"equal, not mixed": {
				createFakeTypeConstraint("@foo"),
				createFakeTypeConstraint("@foo"),
				createFakeTypeConstraint("@foo"),
				"@foo",
			},
			"equal, mixed": {
				createFakeTypeConstraint("mixed"),
				createFakeTypeConstraint("mixed"),
				createFakeTypeConstraint("mixed"),
				"mixed",
			},
			"not equal, new is mixed": {
				createFakeTypeConstraint("@foo"),
				createFakeTypeConstraint("mixed"),
				createFakeTypeConstraint("mixed"),
				"mixed",
			},
		}

		for n, c := range cc {
			t.Run(n, func(t *testing.T) {
				n := createFakeMixedValueNode()
				n.schemaType = "should be changed"
				n.baseNode.AddConstraint(c.exists)

				n.addTypeConstraint(c.new)

				actual := n.baseNode.constraints.GetValue(constraint.TypeConstraintType)
				assert.Equal(t, c.expected, actual)
			})
		}

		t.Run("not equal, new isn't mixed", func(t *testing.T) {
			assert.PanicsWithError(t, `Duplicate "type" rule`, func() {
				n := createFakeMixedValueNode()
				n.schemaType = "should be changed"
				n.baseNode.AddConstraint(createFakeTypeConstraint("@foo"))

				n.addTypeConstraint(createFakeTypeConstraint("@bar"))
			})
		})
	})
}

func Test_addOrConstraint(t *testing.T) {
	t.Run("without type constraint", func(t *testing.T) {
		expected := constraint.NewOr(jschema.RuleASTNodeSourceManual)

		n := createFakeMixedValueNode()
		n.addOrConstraint(expected)

		actual := n.baseNode.constraints.GetValue(constraint.OrConstraintType)
		assert.Equal(t, expected, actual)
	})

	t.Run("with type constraint", func(t *testing.T) {
		expected := constraint.NewOr(jschema.RuleASTNodeSourceManual)

		n := createFakeMixedValueNode()
		n.baseNode.AddConstraint(createFakeTypeConstraint("@foo"))

		n.addOrConstraint(expected)

		actual := n.baseNode.constraints.GetValue(constraint.OrConstraintType)
		assert.Equal(t, expected, actual)

		actual = n.baseNode.constraints.GetValue(constraint.TypeConstraintType)
		assert.Equal(t, createFakeTypeConstraint(`"mixed"`), actual)
	})
}

func TestMixedValueNode_Grow(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		n := createFakeMixedValueNode()
		n.parent = newObjectNode(lexeme.NewLexEvent(lexeme.ObjectBegin, 0, 0, nil))

		cc := map[lexeme.LexEventType]Node{
			lexeme.MixedValueBegin: n,
			lexeme.MixedValueEnd:   n.parent,
		}

		for lexType, expected := range cc {
			t.Run(lexType.String(), func(t *testing.T) {
				actual, ok := n.Grow(lexeme.NewLexEvent(lexType, 0, 0, fs.MustNewFile("", "foo")))
				assert.Equal(t, expected, actual)
				assert.False(t, ok)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithValue(t,
			`Unexpected lexical event "`+lexeme.ObjectBegin.String()+`" in mixed value node`,
			func() {
				createFakeMixedValueNode().
					Grow(lexeme.NewLexEvent(lexeme.ObjectBegin, 0, 0, nil))
			},
		)
	})
}

func createFakeMixedValueNode() *MixedValueNode {
	return NewMixedValueNode(lexeme.NewLexEvent(lexeme.MixedValueBegin, 0, 0, nil))
}

func createFakeTypeConstraint(name string) *constraint.TypeConstraint {
	return constraint.NewType([]byte(name), jschema.RuleASTNodeSourceManual)
}
