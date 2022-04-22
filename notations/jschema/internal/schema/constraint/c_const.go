package constraint

import (
	"j/schema"
	"j/schema/bytes"
	"j/schema/errors"
	"j/schema/internal/json"
	"strconv"
)

type Const struct {
	apply     bool
	nodeValue bytes.Bytes
}

var _ Constraint = Const{}

func NewConst(value, nodeValue bytes.Bytes) *Const {
	c := Const{
		nodeValue: nodeValue,
	}

	var err error
	if c.apply, err = value.ParseBool(); err != nil {
		panic(errors.Format(errors.ErrInvalidValueOfConstraint, ConstType.String()))
	}
	return &c
}

func (Const) IsJsonTypeCompatible(t json.Type) bool {
	return t != json.TypeObject && t != json.TypeArray
}

func (Const) Type() Type {
	return ConstType
}

func (c Const) String() string {
	if c.apply {
		return ConstType.String() + ": true"
	}
	return ConstType.String() + ": false"
}

func (c Const) Bool() bool {
	return c.apply
}

func (c Const) Validate(v bytes.Bytes) {
	if !c.apply {
		return
	}

	if v.String() != c.nodeValue.String() {
		panic(errors.Format(errors.ErrInvalidConst, c.nodeValue.String()))
	}
}

func (c Const) ASTNode() jschema.RuleASTNode {
	return newRuleASTNode(jschema.JSONTypeBoolean, strconv.FormatBool(c.apply), jschema.RuleASTNodeSourceManual)
}
