package checker

import (
	"j/schema/internal/errors"
	"j/schema/internal/lexeme"
	"j/schema/notations/jschema/internal/schema"
)

type arrayChecker struct {
	node schema.Node
}

func newArrayChecker(node schema.Node) arrayChecker {
	return arrayChecker{
		node: node,
	}
}

func (arrayChecker) check(nodeLex lexeme.LexEvent) (err errors.Error) {
	if nodeLex.Type() != lexeme.ArrayEnd {
		return lexeme.NewLexEventError(nodeLex, errors.ErrChecker)
	}

	return nil
}

func (c arrayChecker) indentedString(depth int) string {
	return c.node.IndentedNodeString(depth)
}
