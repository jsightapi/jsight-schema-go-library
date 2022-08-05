package loader

import (
	"fmt"
	"testing"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/internal/lexeme"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/mocks"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema/constraint"

	"github.com/stretchr/testify/assert"
)

func TestSchemaCompiler_precisionConstraint(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]func(*mocks.Node){
			"without constraints": func(n *mocks.Node) {
				n.On("Constraint", constraint.PrecisionConstraintType).Return(nil)
			},
			"with only precision constraint": func(n *mocks.Node) {
				n.On("Constraint", constraint.PrecisionConstraintType).Return(constraint.Precision{})
				n.On("Constraint", constraint.TypeConstraintType).Return(nil)
			},
			"type is decimal": func(n *mocks.Node) {
				n.On("Constraint", constraint.PrecisionConstraintType).Return(constraint.Precision{})
				n.On("Constraint", constraint.TypeConstraintType).Return(constraint.NewType(
					[]byte("decimal"),
					jschema.RuleASTNodeSourceManual,
				))
			},
		}

		for name, fn := range cc {
			t.Run(name, func(t *testing.T) {
				n := mocks.NewNode(t)
				fn(n)
				schemaCompiler{}.precisionConstraint(n)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `The "precision" constraint can't be used for the "foo" type`, func() {
			n := mocks.NewNode(t)
			n.On("Constraint", constraint.PrecisionConstraintType).Return(constraint.Precision{})
			n.On("Constraint", constraint.TypeConstraintType).Return(constraint.NewType(
				[]byte("foo"),
				jschema.RuleASTNodeSourceManual,
			))
			schemaCompiler{}.precisionConstraint(n)
		})
	})
}

func TestSchemaCompiler_emptyArray(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := []schema.Node{
			nil,
			&schema.ObjectNode{},
			func() *schema.ArrayNode {
				n := &schema.ArrayNode{}
				n.Grow(newFakeLexEvent(lexeme.ArrayItemBegin))
				n.Grow(newFakeLexEventWithValue(lexeme.LiteralBegin, "foo"))
				return n
			}(),
			&schema.ArrayNode{},
			func() schema.Node {
				n := schema.NewNode(newFakeLexEvent(lexeme.ArrayBegin))
				n.AddConstraint(constraint.NewMinItems([]byte("0")))
				n.AddConstraint(constraint.NewMaxItems([]byte("0")))
				return n
			}(),
		}

		for i, given := range cc {
			t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
				assert.NotPanics(t, func() {
					schemaCompiler{}.emptyArray(given)
				})
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]schema.Node{
			"min": func() schema.Node {
				n := schema.NewNode(newFakeLexEvent(lexeme.ArrayBegin))
				n.AddConstraint(constraint.NewMinItems([]byte("1")))
				return n
			}(),
			"max": func() schema.Node {
				n := schema.NewNode(newFakeLexEvent(lexeme.ArrayBegin))
				n.AddConstraint(constraint.NewMaxItems([]byte("1")))
				return n
			}(),
		}

		for n, given := range cc {
			t.Run(n, func(t *testing.T) {
				assert.PanicsWithValue(t, errors.ErrIncorrectConstraintValueForEmptyArray, func() {
					schemaCompiler{}.emptyArray(given)
				})
			})
		}
	})
}

func BenchmarkSchemaCompiler_emptyArray(b *testing.B) {
	n := schema.NewNode(newFakeLexEvent(lexeme.ArrayBegin))
	n.AddConstraint(constraint.NewMinItems([]byte("0")))
	n.AddConstraint(constraint.NewMaxItems([]byte("0")))

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		schemaCompiler{}.emptyArray(n)
	}
}
