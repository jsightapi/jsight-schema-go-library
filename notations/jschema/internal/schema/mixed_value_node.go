package schema

import (
	"strings"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
	"github.com/jsightapi/jsight-schema-go-library/internal/lexeme"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema/constraint"
)

type MixedValueNode struct {
	schemaType string
	value      string

	types []string

	baseNode
}

var _ Node = (*MixedValueNode)(nil)

func NewMixedValueNode(lex lexeme.LexEvent) *MixedValueNode {
	n := MixedValueNode{
		baseNode: newBaseNode(lex),
	}
	n.setJsonType(json.TypeMixed)
	n.realType = json.TypeMixed.String()
	return &n
}

func (*MixedValueNode) SetRealType(string) bool {
	// Mixed value node is always have mixed type.
	return true
}

func (n *MixedValueNode) AddConstraint(c constraint.Constraint) {
	switch t := c.(type) {
	case *constraint.TypeConstraint:
		n.addTypeConstraint(t)
		n.types = []string{t.Bytes().String()}

	case *constraint.Or:
		n.addOrConstraint(t)

	case *constraint.TypesList:
		n.types = t.Names()
		n.baseNode.AddConstraint(t)

	default:
		n.baseNode.AddConstraint(t)
	}
}

func (n *MixedValueNode) addTypeConstraint(c *constraint.TypeConstraint) {
	exists, ok := n.constraints.Get(constraint.TypeConstraintType)
	if !ok {
		n.baseNode.AddConstraint(c)
		n.schemaType = c.Bytes().Unquote().String()
		return
	}

	newVal := c.Bytes().Unquote().String()
	existsVal := exists.(constraint.BytesKeeper).Bytes().Unquote().String()
	if newVal != existsVal && newVal != "mixed" {
		panic(errors.Format(errors.ErrDuplicateRule, c.Type().String()))
	}
	n.constraints.Set(c.Type(), c)
	n.schemaType = "mixed"
}

func (n *MixedValueNode) addOrConstraint(c *constraint.Or) {
	if tc, ok := n.constraints.Get(constraint.TypeConstraintType); ok {
		n.addTypeConstraint(constraint.NewType(
			bytes.Bytes(`"mixed"`),
			tc.(*constraint.TypeConstraint).Source(),
		))
	}
	n.baseNode.AddConstraint(c)
}

func (n *MixedValueNode) Grow(lex lexeme.LexEvent) (Node, bool) {
	switch lex.Type() {
	case lexeme.MixedValueBegin:

	case lexeme.MixedValueEnd:
		n.schemaLexEvent = lex
		n.value = lex.Value().TrimSpaces().String()
		n.schemaType = n.value
		return n.parent, false

	default:
		panic(`Unexpected lexical event "` + lex.Type().String() + `" in mixed value node`)
	}

	return n, false
}

func (n *MixedValueNode) ASTNode() (jschema.ASTNode, error) {
	an := astNodeFromNode(n)

	an.SchemaType = n.schemaType
	if strings.ContainsRune(n.value, '|') {
		an.SchemaType = json.TypeMixed.String()
	}
	an.Value = n.value
	return an, nil
}

func (n *MixedValueNode) GetTypes() []string {
	return n.types
}
