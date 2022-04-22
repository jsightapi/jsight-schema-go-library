package constraint

import (
	"j/schema"
	"j/schema/bytes"
	"j/schema/internal/errors"
	"j/schema/internal/json"
)

type MinItems struct {
	rawValue bytes.Bytes
	value    *json.Number
}

var _ Constraint = MinItems{}

func NewMinItems(ruleValue bytes.Bytes) *MinItems {
	number, err := json.NewIntegerNumber(ruleValue)
	if err != nil {
		panic(err)
	}

	return &MinItems{
		rawValue: ruleValue,
		value:    number,
	}
}

func (MinItems) IsJsonTypeCompatible(t json.Type) bool {
	return t == json.TypeArray
}

func (MinItems) Type() Type {
	return MinItemsConstraintType
}

func (c MinItems) String() string {
	return MinItemsConstraintType.String() + ": " + c.value.String()
}

func (c MinItems) ValidateTheArray(numberOfChildren uint) {
	length := json.NewNumberFromUint(numberOfChildren)
	if length.LessThan(c.value) {
		panic(errors.ErrConstraintMinItemsValidation)
	}
}

func (c MinItems) Value() *json.Number {
	return c.value
}

func (c MinItems) ASTNode() jschema.RuleASTNode {
	return newRuleASTNode(jschema.JSONTypeNumber, c.rawValue.String(), jschema.RuleASTNodeSourceManual)
}
