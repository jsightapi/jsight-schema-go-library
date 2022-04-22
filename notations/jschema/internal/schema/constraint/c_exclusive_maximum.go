package constraint //nolint:dupl // Duplicates exclusive minimum with small differences.

import (
	"j/schema"
	"j/schema/bytes"
	"j/schema/internal/errors"
	"j/schema/internal/json"
	"strconv"
)

type ExclusiveMaximum struct {
	exclusive bool
}

var _ Constraint = ExclusiveMaximum{}

func NewExclusiveMaximum(ruleValue bytes.Bytes) *ExclusiveMaximum {
	c := ExclusiveMaximum{}
	var err error
	if c.exclusive, err = ruleValue.ParseBool(); err != nil {
		panic(errors.Format(errors.ErrInvalidValueOfConstraint, ExclusiveMaximumConstraintType.String()))
	}
	return &c
}

func (ExclusiveMaximum) IsJsonTypeCompatible(t json.Type) bool {
	if t == json.TypeInteger || t == json.TypeFloat {
		return true
	}
	return false
}

func (ExclusiveMaximum) Type() Type {
	return ExclusiveMaximumConstraintType
}

func (c ExclusiveMaximum) String() string {
	str := "UNVERIFIABLE CONSTRAINT " + ExclusiveMaximumConstraintType.String()
	if c.exclusive {
		str += ": true"
	} else {
		str += ": false"
	}
	return str
}

func (c ExclusiveMaximum) IsExclusive() bool {
	return c.exclusive
}

func (c ExclusiveMaximum) ASTNode() jschema.RuleASTNode {
	return newRuleASTNode(jschema.JSONTypeBoolean, strconv.FormatBool(c.exclusive), jschema.RuleASTNodeSourceManual)
}
