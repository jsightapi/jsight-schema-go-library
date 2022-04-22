package constraint

import (
	"j/schema"
	"j/schema/internal/json"
)

type AnyConstraint struct{}

var _ Constraint = AnyConstraint{}

func NewAny() *AnyConstraint {
	return &AnyConstraint{}
}

func (AnyConstraint) IsJsonTypeCompatible(json.Type) bool {
	return true
}

func (AnyConstraint) Type() Type {
	return AnyConstraintType
}

func (AnyConstraint) String() string {
	return AnyConstraintType.String()
}

func (AnyConstraint) ASTNode() jschema.RuleASTNode {
	return newEmptyRuleASTNode()
}
