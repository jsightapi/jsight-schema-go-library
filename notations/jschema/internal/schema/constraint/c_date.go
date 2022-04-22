package constraint

import (
	"j/schema"
	"j/schema/bytes"
	"j/schema/errors"
	"j/schema/internal/json"
	"time"
)

type Date struct {
}

var _ Constraint = Date{}

func NewDate() *Date {
	return &Date{}
}

func (Date) IsJsonTypeCompatible(t json.Type) bool {
	return t == json.TypeString
}

func (Date) Type() Type {
	return DateConstraintType
}

func (Date) String() string {
	return DateConstraintType.String()
}

func (Date) Validate(value bytes.Bytes) {
	str := value.Unquote().String()
	_, err := time.Parse("2006-01-02", str)
	if err != nil {
		panic(errors.Format(errors.ErrInvalidDate, err))
	}
}

func (Date) ASTNode() jschema.RuleASTNode {
	return newEmptyRuleASTNode()
}
