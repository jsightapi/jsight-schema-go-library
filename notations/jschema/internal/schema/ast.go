package schema

import (
	"j/schema"
	"j/schema/notations/jschema/internal/schema/constraint"
)

func newASTNode() jschema.ASTNode {
	return jschema.ASTNode{
		Rules:      &jschema.RuleASTNodes{},
		Properties: &jschema.ASTNodes{},
	}
}

func astNodeFromNode(n Node) jschema.ASTNode {
	an := newASTNode()

	an.JSONType = n.Type().ToJSONType()
	an.SchemaType = getASTNodeSchemaType(n)
	an.Rules = collectASTRules(n.ConstraintMap())
	an.Comment = n.Comment()

	return an
}

func getASTNodeSchemaType(n Node) string {
	if n.Constraint(constraint.EnumConstraintType) != nil {
		return "enum"
	}

	if n.Constraint(constraint.OrConstraintType) != nil {
		return jschema.JSONTypeMixed
	}

	if c := n.Constraint(constraint.TypeConstraintType); c != nil {
		if tc, ok := c.(*constraint.TypeConstraint); ok {
			return tc.Bytes().Unquote().String()
		}
	}

	if n.Constraint(constraint.PrecisionConstraintType) != nil {
		return "decimal"
	}

	return n.Type().String()
}

func collectASTRules(cc *Constraints) *jschema.RuleASTNodes {
	nn := &jschema.RuleASTNodes{}

	for kv := range cc.Iterate() {
		switch kv.Key {
		// The `Or` constraint doesn't contain all required values, but they are placed
		// in the `type` constraint.
		case constraint.OrConstraintType:
			types, ok := cc.Get(constraint.TypesListConstraintType)
			if !ok {
				panic(`Can't collect rules: "types" constraint is required with "or"" constraint`)
			}

			nn.Set(constraint.OrConstraintType.String(), types.ASTNode())

		case constraint.TypesListConstraintType:
			// do nothing

		default:
			nn.Set(kv.Key.String(), kv.Value.ASTNode())
		}
	}
	return nn
}
