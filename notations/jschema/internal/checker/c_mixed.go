package checker

import (
	"fmt"
	"j/schema/errors"
	"j/schema/internal/lexeme"
	"j/schema/notations/jschema/internal/schema"
	"j/schema/notations/jschema/internal/validator"
)

type mixedChecker struct {
	node schema.Node
}

func newMixedChecker(node schema.Node) mixedChecker {
	return mixedChecker{
		node: node,
	}
}

func (c mixedChecker) check(nodeLex lexeme.LexEvent) (err errors.Error) {
	defer func() {
		if r := recover(); r != nil {
			switch val := r.(type) {
			case errors.DocumentError:
				err = val
			case errors.Err:
				err = lexeme.NewLexEventError(nodeLex, val)
			default:
				err = lexeme.NewLexEventError(nodeLex, errors.Format(errors.ErrGeneric, fmt.Sprintf("%s", r)))
			}
		}
	}()

	if nodeLex.Type() == lexeme.LiteralEnd {
		validator.ValidateLiteralValue(c.node, nodeLex.Value()) // can panic
	}

	return nil
}

func (c mixedChecker) indentedString(depth int) string {
	return c.node.IndentedNodeString(depth)
}
