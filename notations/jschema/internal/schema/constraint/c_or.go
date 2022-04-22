package constraint

import (
	"j/schema"
	"j/schema/internal/json"
)

// Used for compile-time checking.

type Or struct {
	source jschema.RuleASTNodeSource
}

var _ Constraint = Or{}

func NewOr(s jschema.RuleASTNodeSource) *Or {
	return &Or{
		source: s,
	}
}

func (c Or) IsGenerated() bool {
	return c.source == jschema.RuleASTNodeSourceGenerated
}

func (Or) IsJsonTypeCompatible(json.Type) bool {
	return true
}

func (Or) Type() Type {
	return OrConstraintType
}

func (Or) String() string {
	return "[ UNVERIFIABLE CONSTRAINT ] " + OrConstraintType.String()
}

func (Or) ASTNode() jschema.RuleASTNode {
	// Check `collectASTRules` function for the actual logic.
	return newEmptyRuleASTNode()
}
