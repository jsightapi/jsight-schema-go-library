package checker

import (
	"j/schema/internal/errors"
	"j/schema/notations/jschema/internal/schema"
	"j/schema/notations/jschema/internal/schema/constraint"
)

type nodeCheckerListConstructor struct {
	// A rootSchema from which it is possible to receive type by their name.
	rootSchema *schema.Schema

	// types the map of used types.
	types map[string]schema.Type

	// A list of checkers for the node.
	list []nodeChecker

	// addedTypeNames a set of already added types. Exists for excluding
	// recursive addition of type to the list.
	addedTypeNames map[string]bool
}

func (l *nodeCheckerListConstructor) buildList(node schema.Node) {
	constr := node.Constraint(constraint.TypesListConstraintType)
	if constr != nil {
		names := constr.(*constraint.TypesList).Names()
		l.appendTypeValidators(names)
	} else {
		l.appendNodeValidators(node)
	}
}

func (l *nodeCheckerListConstructor) appendTypeValidators(names []string) {
	if l.list == nil {
		l.addedTypeNames = make(map[string]bool, len(names)) // optimizing memory allocation
		l.list = make([]nodeChecker, 0, len(names))          // optimizing memory allocation
	}
	for _, name := range names {
		if _, ok := l.addedTypeNames[name]; !ok {
			l.addedTypeNames[name] = true
			l.buildList(getType(name, l.rootSchema, l.types).RootNode()) // can panic
		}
	}
}

func (l *nodeCheckerListConstructor) appendNodeValidators(node schema.Node) {
	if l.list == nil {
		l.list = make([]nodeChecker, 0, 1) // optimizing memory allocation
	}

	var c nodeChecker

	switch node.(type) {
	case *schema.LiteralNode:
		c = newLiteralChecker(node)
	case *schema.ObjectNode:
		c = newObjectChecker(node)
	case *schema.ArrayNode:
		c = newArrayChecker(node)
	case *schema.MixedNode:
		c = newMixedChecker(node)
	default:
		panic(errors.ErrImpossible)
	}

	l.list = append(l.list, c)
}
