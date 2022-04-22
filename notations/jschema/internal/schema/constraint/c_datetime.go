package constraint

import (
	"j/schema"
	"j/schema/bytes"
	"j/schema/internal/errors"
	"j/schema/internal/json"
	"time"
)

type DateTime struct{}

var _ Constraint = DateTime{}

func NewDateTime() *DateTime {
	return &DateTime{}
}

func (DateTime) IsJsonTypeCompatible(t json.Type) bool {
	return t == json.TypeString
}

func (DateTime) Type() Type {
	return DateTimeConstraintType
}

func (DateTime) String() string {
	return DateTimeConstraintType.String()
}

func (DateTime) Validate(value bytes.Bytes) {
	str := value.Unquote().String()
	_, err := time.Parse(time.RFC3339, str)
	if err != nil {
		panic(errors.ErrInvalidDateTime)
	}
}

func (DateTime) ASTNode() jschema.RuleASTNode {
	return newEmptyRuleASTNode()
}
