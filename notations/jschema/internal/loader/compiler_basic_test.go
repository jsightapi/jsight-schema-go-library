package loader

import (
	jschema "j/schema"
	"testing"

	"j/schema/notations/jschema/internal/mocks"
	"j/schema/notations/jschema/internal/schema/constraint"

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
				n := &mocks.Node{}
				fn(n)
				schemaCompiler{}.precisionConstraint(n)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `The "precision" constraint can't be used for the "foo" type`, func() {
			n := &mocks.Node{}
			n.On("Constraint", constraint.PrecisionConstraintType).Return(constraint.Precision{})
			n.On("Constraint", constraint.TypeConstraintType).Return(constraint.NewType(
				[]byte("foo"),
				jschema.RuleASTNodeSourceManual,
			))
			schemaCompiler{}.precisionConstraint(n)
		})
	})
}
