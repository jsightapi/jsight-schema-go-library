package constraint

import (
	"j/schema"
	"j/schema/bytes"
	"j/schema/internal/errors"
	"j/schema/internal/json"
	"net/url"
)

type Uri struct{}

var _ Constraint = Uri{}

func NewUri() *Uri {
	return &Uri{}
}

func (Uri) IsJsonTypeCompatible(t json.Type) bool {
	return t == json.TypeString
}

func (Uri) Type() Type {
	return UriConstraintType
}

func (Uri) String() string {
	return UriConstraintType.String()
}

func (Uri) Validate(value bytes.Bytes) {
	u, err := url.ParseRequestURI(value.Unquote().String())
	if err != nil {
		panic(errors.Format(errors.ErrInvalidUri, value.Unquote().String()))
	}
	if !u.IsAbs() {
		panic(errors.Format(errors.ErrInvalidUri, value.Unquote().String()))
	}
	if u.Hostname() == "" {
		panic(errors.Format(errors.ErrInvalidUri, value.Unquote().String()))
	}
}

func (Uri) ASTNode() jschema.RuleASTNode {
	return newEmptyRuleASTNode()
}
