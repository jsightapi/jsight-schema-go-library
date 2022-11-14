package checker

import (
	"fmt"

	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/lexeme"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/ischema"
)

type mixedChecker struct {
	node ischema.Node
}

func newMixedChecker(node ischema.Node) mixedChecker {
	return mixedChecker{
		node: node,
	}
}

func (c mixedChecker) Check(nodeLex lexeme.LexEvent) (err errors.Error) {
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
		ValidateLiteralValue(c.node, nodeLex.Value()) // can panic
	}

	return nil
}
