package schema

import (
	"j/schema"
	"j/schema/errors"
	"j/schema/internal/json"
	"j/schema/internal/lexeme"
	"strings"
)

type MixedNode struct {
	baseNode
	// mixed bool // for debug
}

var _ Node = &MixedNode{}

func NewMixedNode(lex lexeme.LexEvent) *MixedNode {
	n := MixedNode{
		baseNode: newBaseNode(lex),
		// mixed: true,
	}
	n.setJsonType(json.Guess(lex.Value()).JsonType())
	return &n
}

// for mixed node n.baseNode.jsonType is an EXAMPLE type
func (n *MixedNode) SetJsonType(t json.Type) {
	n.setJsonType(t)
}

func (*MixedNode) Grow(lexeme.LexEvent) (Node, bool) {
	panic(errors.ErrNodeGrow)
}

func (n MixedNode) IndentedTreeString(depth int) string {
	return n.IndentedNodeString(depth)
}

func (n MixedNode) IndentedNodeString(depth int) string {
	indent := strings.Repeat("\t", depth)

	var str strings.Builder
	str.WriteString(indent + "* " + n.Type().String() + "\n")

	for kv := range n.constraints.Iterate() {
		str.WriteString(indent + "* " + kv.Value.String() + "\n")
	}

	return str.String()
}

func (n MixedNode) ASTNode() (jschema.ASTNode, error) { // todo fix
	an := newASTNode()

	an.SchemaType = n.Type().String()
	an.Value = n.Value().Unquote().String()
	an.Rules = collectASTRules(n.constraints)
	an.Comment = n.comment

	return an, nil
}
