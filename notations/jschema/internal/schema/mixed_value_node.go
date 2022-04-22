package schema

import (
	"j/schema"
	"j/schema/bytes"
	"j/schema/errors"
	"j/schema/internal/json"
	"j/schema/internal/lexeme"
	"j/schema/notations/jschema/internal/schema/constraint"
	"strings"
)

type MixedValueNode struct {
	baseNode

	schemaType string
	value      string
}

var _ Node = &MixedValueNode{}

func NewMixedValueNode(lex lexeme.LexEvent) *MixedValueNode {
	n := MixedValueNode{
		baseNode: newBaseNode(lex),
	}
	n.setJsonType(json.TypeMixed)
	return &n
}

func (n *MixedValueNode) AddConstraint(c constraint.Constraint) {
	switch t := c.(type) {
	case *constraint.TypeConstraint:
		n.addTypeConstraint(t)

	case *constraint.Or:
		n.addOrConstraint(t)

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

func (n *MixedValueNode) IndentedTreeString(depth int) string {
	return n.IndentedNodeString(depth)
}

func (n *MixedValueNode) IndentedNodeString(depth int) string {
	indent := strings.Repeat("\t", depth)

	var str strings.Builder
	str.WriteString(indent + "* " + n.Type().String() + "\n")

	for kv := range n.constraints.Iterate() {
		str.WriteString(indent + "* " + kv.Value.String() + "\n")
	}

	return str.String()
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
