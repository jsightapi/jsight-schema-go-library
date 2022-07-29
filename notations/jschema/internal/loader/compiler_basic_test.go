package loader

import (
	"testing"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/mocks"
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
