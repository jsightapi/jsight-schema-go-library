package constraint

import (
	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

type TypeConstraint struct {
	value  bytes.Bytes
	source jschema.RuleASTNodeSource
}

var (
	_ Constraint  = TypeConstraint{}
	_ Constraint  = (*TypeConstraint)(nil)
	_ BytesKeeper = TypeConstraint{}
	_ BytesKeeper = (*TypeConstraint)(nil)
)

func NewType(ruleValue bytes.Bytes, source jschema.RuleASTNodeSource) *TypeConstraint {
	return &TypeConstraint{
		value:  ruleValue,
		source: source,
	}
}

func (c TypeConstraint) IsGenerated() bool {
	return c.source == jschema.RuleASTNodeSourceGenerated
}

func (TypeConstraint) IsJsonTypeCompatible(json.Type) bool {
	return true
}

func (TypeConstraint) Type() Type {
	return TypeConstraintType
}

func (c TypeConstraint) String() string {
	return TypeConstraintType.String() + ": " + c.value.String()
}

func (c TypeConstraint) Bytes() bytes.Bytes {
	return c.value
}

func (c TypeConstraint) ASTNode() jschema.RuleASTNode {
	t := jschema.TokenTypeString
	if c.value.Unquote().IsUserTypeName() {
		t = jschema.TokenTypeShortcut
	}
	return newRuleASTNode(t, c.value.Unquote().String(), c.source)
}

func (c TypeConstraint) Source() jschema.RuleASTNodeSource { return c.source }
