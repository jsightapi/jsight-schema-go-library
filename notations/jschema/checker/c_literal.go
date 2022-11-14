package checker

import (
	"fmt"

	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/lexeme"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/ischema"
)

type literalChecker struct {
	node ischema.Node
}

func newLiteralChecker(node ischema.Node) literalChecker {
	return literalChecker{
		node: node,
	}
}

func (c literalChecker) Check(nodeLex lexeme.LexEvent) (err errors.Error) {
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

	if nodeLex.Type() != lexeme.LiteralEnd {
		return lexeme.NewLexEventError(nodeLex, errors.ErrChecker)
	}

	ValidateLiteralValue(c.node, nodeLex.Value()) // can panic

	return nil
}
