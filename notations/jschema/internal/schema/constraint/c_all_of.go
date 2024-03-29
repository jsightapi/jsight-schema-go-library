package constraint

import (
	"strings"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

type AllOf struct {
	schemaName []string
}

var (
	_ Constraint = AllOf{}
	_ Constraint = (*AllOf)(nil)
)

func NewAllOf() *AllOf {
	return &AllOf{
		schemaName: make([]string, 0, 3),
	}
}

func (AllOf) IsJsonTypeCompatible(t json.Type) bool {
	return t == json.TypeObject
}

func (AllOf) Type() Type {
	return AllOfConstraintType
}

func (c AllOf) String() string {
	return AllOfConstraintType.String() + ": " + strings.Join(c.schemaName, ", ")
}

func (c *AllOf) Append(scalar bytes.Bytes) {
	if !json.Guess(scalar).IsString() {
		panic(errors.ErrUnacceptableValueInAllOfRule)
	}

	s := scalar.Unquote()

	if !s.IsUserTypeName() {
		panic(errors.Format(errors.ErrInvalidSchemaNameInAllOfRule, s))
	}
	c.schemaName = append(c.schemaName, s.String())
}

func (c AllOf) SchemaNames() []string {
	return c.schemaName
}

func (c AllOf) ASTNode() jschema.RuleASTNode {
	const source = jschema.RuleASTNodeSourceManual

	if len(c.schemaName) == 1 {
		return newRuleASTNode(jschema.TokenTypeShortcut, c.schemaName[0], source)
	}

	n := newRuleASTNode(jschema.TokenTypeArray, "", source)
	n.Items = make([]jschema.RuleASTNode, 0, len(c.schemaName))

	for _, sn := range c.schemaName {
		n.Items = append(n.Items, newRuleASTNode(jschema.TokenTypeShortcut, sn, source))
	}

	return n
}
