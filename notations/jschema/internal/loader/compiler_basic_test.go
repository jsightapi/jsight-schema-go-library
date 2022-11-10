package loader

import (
	"fmt"
	"testing"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/internal/lexeme"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/mocks"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/schema"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/schema/constraint"

	"github.com/stretchr/testify/assert"
)

func TestSchemaCompiler_checkMinAndMax(t *testing.T) {
	cc := map[string]struct {
		node        func(*testing.T) schema.Node
		expectedErr string
	}{
		"nil min, nil max": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinConstraintType).Return(nil)
				m.On("Constraint", constraint.MaxConstraintType).Return(nil)
				return m
			},
		},
		"nil min, not nil max": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinConstraintType).Return(nil)
				m.On("Constraint", constraint.MaxConstraintType).Return(constraint.NewMax(bytes.NewBytes("42")))
				return m
			},
		},
		"nil min, not nil max (exclusive)": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				max := constraint.NewMax(bytes.NewBytes("42"))
				max.SetExclusive(true)
				m.On("Constraint", constraint.MinConstraintType).Return(nil)
				m.On("Constraint", constraint.MaxConstraintType).Return(max)
				return m
			},
		},
		"not nil min, nil max": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinConstraintType).Return(constraint.NewMin(bytes.NewBytes("42")))
				m.On("Constraint", constraint.MaxConstraintType).Return(nil)
				return m
			},
		},
		"not nil min (exclusive), nil max": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				min := constraint.NewMin(bytes.NewBytes("42"))
				min.SetExclusive(true)
				m.On("Constraint", constraint.MinConstraintType).Return(min)
				m.On("Constraint", constraint.MaxConstraintType).Return(nil)
				return m
			},
		},
		"min < max": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinConstraintType).Return(constraint.NewMin(bytes.NewBytes("1")))
				m.On("Constraint", constraint.MaxConstraintType).Return(constraint.NewMax(bytes.NewBytes("2")))
				return m
			},
		},
		"min (exclusive) < max": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				min := constraint.NewMin(bytes.NewBytes("1"))
				min.SetExclusive(true)
				m.On("Constraint", constraint.MinConstraintType).Return(min)
				m.On("Constraint", constraint.MaxConstraintType).Return(constraint.NewMax(bytes.NewBytes("2")))
				return m
			},
		},
		"min < max (exclusive)": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				max := constraint.NewMax(bytes.NewBytes("2"))
				max.SetExclusive(true)
				m.On("Constraint", constraint.MinConstraintType).Return(constraint.NewMin(bytes.NewBytes("1")))
				m.On("Constraint", constraint.MaxConstraintType).Return(max)
				return m
			},
		},
		"min (exclusive) < max (exclusive)": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				min := constraint.NewMin(bytes.NewBytes("1"))
				min.SetExclusive(true)
				max := constraint.NewMax(bytes.NewBytes("2"))
				max.SetExclusive(true)
				m.On("Constraint", constraint.MinConstraintType).Return(min)
				m.On("Constraint", constraint.MaxConstraintType).Return(max)
				return m
			},
		},
		"min = max": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinConstraintType).Return(constraint.NewMin(bytes.NewBytes("1")))
				m.On("Constraint", constraint.MaxConstraintType).Return(constraint.NewMax(bytes.NewBytes("1")))
				return m
			},
		},
		"min (exclusive) = max": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				min := constraint.NewMin(bytes.NewBytes("1"))
				min.SetExclusive(true)
				m.On("Constraint", constraint.MinConstraintType).Return(min)
				m.On("Constraint", constraint.MaxConstraintType).Return(constraint.NewMax(bytes.NewBytes("1")))
				return m
			},
			expectedErr: `Value of constraint "min" should be less than value of "max" constraint`,
		},
		"min = max (exclusive)": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				max := constraint.NewMax(bytes.NewBytes("1"))
				max.SetExclusive(true)
				m.On("Constraint", constraint.MinConstraintType).Return(constraint.NewMin(bytes.NewBytes("1")))
				m.On("Constraint", constraint.MaxConstraintType).Return(max)
				return m
			},
			expectedErr: `Value of constraint "min" should be less than value of "max" constraint`,
		},
		"min (exclusive) = max (exclusive)": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				min := constraint.NewMin(bytes.NewBytes("1"))
				min.SetExclusive(true)
				max := constraint.NewMax(bytes.NewBytes("1"))
				max.SetExclusive(true)
				m.On("Constraint", constraint.MinConstraintType).Return(min)
				m.On("Constraint", constraint.MaxConstraintType).Return(max)
				return m
			},
			expectedErr: `Value of constraint "min" should be less than value of "max" constraint`,
		},
		"min > max": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinConstraintType).Return(constraint.NewMin(bytes.NewBytes("2")))
				m.On("Constraint", constraint.MaxConstraintType).Return(constraint.NewMax(bytes.NewBytes("1")))
				return m
			},
			expectedErr: `Value of constraint "min" should be less or equal to value of "max" constraint`,
		},
		"min (exclusive) > max": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				min := constraint.NewMin(bytes.NewBytes("2"))
				min.SetExclusive(true)
				m.On("Constraint", constraint.MinConstraintType).Return(min)
				m.On("Constraint", constraint.MaxConstraintType).Return(constraint.NewMax(bytes.NewBytes("1")))
				return m
			},
			expectedErr: `Value of constraint "min" should be less than value of "max" constraint`,
		},
		"min > max (exclusive)": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				max := constraint.NewMax(bytes.NewBytes("1"))
				max.SetExclusive(true)
				m.On("Constraint", constraint.MinConstraintType).Return(constraint.NewMin(bytes.NewBytes("2")))
				m.On("Constraint", constraint.MaxConstraintType).Return(max)
				return m
			},
			expectedErr: `Value of constraint "min" should be less than value of "max" constraint`,
		},
		"min (exclusive) > max (exclusive)": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				min := constraint.NewMin(bytes.NewBytes("2"))
				min.SetExclusive(true)
				max := constraint.NewMax(bytes.NewBytes("1"))
				max.SetExclusive(true)
				m.On("Constraint", constraint.MinConstraintType).Return(min)
				m.On("Constraint", constraint.MaxConstraintType).Return(max)
				return m
			},
			expectedErr: `Value of constraint "min" should be less than value of "max" constraint`,
		},
	}

	for n, c := range cc {
		t.Run(n, func(t *testing.T) {
			err := schemaCompiler{}.checkMinAndMax(c.node(t))
			if c.expectedErr != "" {
				assert.EqualError(t, err, c.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSchemaCompiler_checkMinLengthAndMaxLength(t *testing.T) {
	cc := map[string]struct {
		node        func(*testing.T) schema.Node
		expectedErr string
	}{
		"nil minLength, nil maxLength": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinLengthConstraintType).Return(nil)
				m.On("Constraint", constraint.MaxLengthConstraintType).Return(nil)
				return m
			},
		},
		"nil minLength, not nil maxLength": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinLengthConstraintType).Return(nil)
				m.On("Constraint", constraint.MaxLengthConstraintType).Return(constraint.NewMaxLength(bytes.NewBytes("42")))
				return m
			},
		},
		"not nil minLength, nil maxLength": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinLengthConstraintType).Return(constraint.NewMinLength(bytes.NewBytes("42")))
				m.On("Constraint", constraint.MaxLengthConstraintType).Return(nil)
				return m
			},
		},
		"minLength < maxLength": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinLengthConstraintType).Return(constraint.NewMinLength(bytes.NewBytes("1")))
				m.On("Constraint", constraint.MaxLengthConstraintType).Return(constraint.NewMaxLength(bytes.NewBytes("2")))
				return m
			},
		},
		"minLength = maxLength": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinLengthConstraintType).Return(constraint.NewMinLength(bytes.NewBytes("2")))
				m.On("Constraint", constraint.MaxLengthConstraintType).Return(constraint.NewMaxLength(bytes.NewBytes("2")))
				return m
			},
		},
		"minLength > maxLength": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinLengthConstraintType).Return(constraint.NewMinLength(bytes.NewBytes("2")))
				m.On("Constraint", constraint.MaxLengthConstraintType).Return(constraint.NewMaxLength(bytes.NewBytes("1")))
				return m
			},
			expectedErr: `Value of constraint "minLength" should be less or equal to value of "maxLength" constraint`,
		},
	}

	for n, c := range cc {
		t.Run(n, func(t *testing.T) {
			err := schemaCompiler{}.checkMinLengthAndMaxLength(c.node(t))
			if c.expectedErr != "" {
				assert.EqualError(t, err, c.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSchemaCompiler_checkMinItemsAndMaxItems(t *testing.T) {
	cc := map[string]struct {
		node        func(*testing.T) schema.Node
		expectedErr string
	}{
		"nil minItems, nil maxItems": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinItemsConstraintType).Return(nil)
				m.On("Constraint", constraint.MaxItemsConstraintType).Return(nil)
				return m
			},
		},
		"nil minItems, not nil maxItems": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinItemsConstraintType).Return(nil)
				m.On("Constraint", constraint.MaxItemsConstraintType).Return(constraint.NewMaxItems(bytes.NewBytes("42")))
				return m
			},
		},
		"not nil minItems, nil maxItems": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinItemsConstraintType).Return(constraint.NewMinItems(bytes.NewBytes("42")))
				m.On("Constraint", constraint.MaxItemsConstraintType).Return(nil)
				return m
			},
		},
		"minItems < maxItems": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinItemsConstraintType).Return(constraint.NewMinItems(bytes.NewBytes("1")))
				m.On("Constraint", constraint.MaxItemsConstraintType).Return(constraint.NewMaxItems(bytes.NewBytes("2")))
				return m
			},
		},
		"minItems = maxItems": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinItemsConstraintType).Return(constraint.NewMinItems(bytes.NewBytes("2")))
				m.On("Constraint", constraint.MaxItemsConstraintType).Return(constraint.NewMaxItems(bytes.NewBytes("2")))
				return m
			},
		},
		"minItems > maxItems": {
			node: func(t *testing.T) schema.Node {
				m := mocks.NewNode(t)
				m.On("Constraint", constraint.MinItemsConstraintType).Return(constraint.NewMinItems(bytes.NewBytes("2")))
				m.On("Constraint", constraint.MaxItemsConstraintType).Return(constraint.NewMaxItems(bytes.NewBytes("1")))
				return m
			},
			expectedErr: `Value of constraint "minItems" should be less or equal to value of "maxItems" constraint`,
		},
	}

	for n, c := range cc {
		t.Run(n, func(t *testing.T) {
			err := schemaCompiler{}.checkMinItemsAndMaxItems(c.node(t))
			if c.expectedErr != "" {
				assert.EqualError(t, err, c.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

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
					bytes.NewBytes("decimal"),
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
				bytes.NewBytes("foo"),
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
				n.AddConstraint(constraint.NewMinItems(bytes.NewBytes("0")))
				n.AddConstraint(constraint.NewMaxItems(bytes.NewBytes("0")))
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
				n.AddConstraint(constraint.NewMinItems(bytes.NewBytes("1")))
				return n
			}(),
			"max": func() schema.Node {
				n := schema.NewNode(newFakeLexEvent(lexeme.ArrayBegin))
				n.AddConstraint(constraint.NewMaxItems(bytes.NewBytes("1")))
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
	n.AddConstraint(constraint.NewMinItems(bytes.NewBytes("0")))
	n.AddConstraint(constraint.NewMaxItems(bytes.NewBytes("0")))

	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		schemaCompiler{}.emptyArray(n)
	}
}
