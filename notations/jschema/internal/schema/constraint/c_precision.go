package constraint

import (
	"j/schema"
	"j/schema/bytes"
	"j/schema/internal/errors"
	"j/schema/internal/json"
	"strconv"
)

type Precision struct {
	value uint
}

var _ Constraint = Precision{}

func NewPrecision(ruleValue bytes.Bytes) *Precision {
	c := Precision{}

	u, err := ruleValue.ParseUint()
	if err != nil {
		panic(errors.Format(errors.ErrInvalidValueOfConstraint, PrecisionConstraintType.String()))
	}

	if u == 0 {
		panic(errors.ErrZeroPrecision)
	}

	c.value = u
	return &c
}

func (Precision) IsJsonTypeCompatible(t json.Type) bool {
	return t == json.TypeFloat
}

func (Precision) Type() Type {
	return PrecisionConstraintType
}

func (c Precision) String() string {
	return PrecisionConstraintType.String() + ": " + strconv.Itoa(int(c.value))
}

func (c Precision) Validate(value bytes.Bytes) {
	n, err := json.NewNumber(value)
	if err != nil {
		panic(err)
	}
	if c.value < n.LengthOfFractionalPart() {
		panic(errors.Format(
			errors.ErrConstraintValidation,
			PrecisionConstraintType.String(),
			strconv.Itoa(int(c.value)),
			"(exclusive)",
		))
	}
}

func (c Precision) ASTNode() jschema.RuleASTNode {
	return newRuleASTNode(jschema.JSONTypeNumber, strconv.FormatUint(uint64(c.value), 10), jschema.RuleASTNodeSourceManual)
}
