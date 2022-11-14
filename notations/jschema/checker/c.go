package checker

import (
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/lexeme"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/ischema"
)

type nodeChecker interface {
	Check(lexeme.LexEvent) errors.Error
}

func newNodeChecker(node ischema.Node) (nodeChecker, error) {
	switch node.(type) {
	case *ischema.LiteralNode:
		return newLiteralChecker(node), nil

	case *ischema.ObjectNode:
		return newObjectChecker(), nil

	case *ischema.ArrayNode:
		return newArrayChecker(), nil

	case *ischema.MixedNode:
		return newMixedChecker(node), nil

	default:
		return nil, errors.ErrImpossible
	}
}
