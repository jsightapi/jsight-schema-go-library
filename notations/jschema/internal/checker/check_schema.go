package checker

import (
	"j/schema/errors"
	"j/schema/internal/json"
	"j/schema/internal/lexeme"
	"j/schema/internal/logger"
	"j/schema/notations/jschema/internal/schema"
	"j/schema/notations/jschema/internal/schema/constraint"
)

// Checks the SAMPLE SCHEMA and all TYPES for compliance with all RULES.

type checkSchema struct {
	rootSchema *schema.Schema

	// foundTypeNames the names of the type encountered during checking. Are used
	// to control recursion.
	foundTypeNames map[string]bool

	// allowedJsonTypes the list of available json-types from types.
	allowedJsonTypes map[json.Type]bool
	log              logger.Logger
}

func CheckRootSchema(rootSchema *schema.Schema, log logger.Logger) {
	c := checkSchema{
		rootSchema:       rootSchema,
		foundTypeNames:   make(map[string]bool, 10),
		allowedJsonTypes: make(map[json.Type]bool, 10),
		log:              log,
	}

	if rootSchema.RootNode() != nil { // the root schema may contain no nodes
		log.Info("--------------------")
		log.Info("Checking ROOT-SCHEMA...\n")
		c.checkNode(rootSchema.RootNode(), nil)
	}

	for name, typ := range rootSchema.TypesList() {
		log.Info("--------------------")
		log.Info("Checking TYPE (" + name + ")...\n")
		c.checkType(name, typ)
	}

	log.Info("ROOT-SCHEMA is checked")
}

func (c *checkSchema) checkType(name string, typ schema.Type) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		if documentError, ok := r.(errors.DocumentError); ok { // return an error with the full set of bytes of the root schema
			documentError.SetFile(typ.RootFile())
			documentError.SetIndex(documentError.Index() + typ.Begin())
			documentError.SetIncorrectUserType(name)
			panic(documentError)
		}

		panic(r)
	}()

	c.checkNode(typ.Schema().RootNode(), typ.Schema().TypesList())
}

func (c *checkSchema) checkerList(node schema.Node, ss map[string]schema.Type) []nodeChecker {
	l := nodeCheckerListConstructor{
		rootSchema: c.rootSchema,
		types:      ss,
	}
	l.buildList(node)
	return l.list
}

func (c checkSchema) checkNode(node schema.Node, ss map[string]schema.Type) {
	defer lexeme.CatchLexEventError(node.BasisLexEventOfSchemaForNode())
	switch node := node.(type) {
	case *schema.LiteralNode:
		c.log.Notice("Check literal node:")
		c.log.Default("value = " + node.BasisLexEventOfSchemaForNode().Value().String() + "\n")
		// c.log.Default(node.IndentedNodeString(0))
		c.checkCompatibilityOfConstraints(node)
		c.checkLinksOfNode(node, ss) // can panic
		c.checkLiteralNode(node, ss)
	case *schema.ArrayNode:
		c.log.Notice("Check array node:")
		c.log.Default(node.IndentedNodeString(0))
		c.checkCompatibilityOfConstraints(node)
		c.checkLinksOfNode(node, ss) // can panic
		c.checkArrayItems(node)
	case *schema.ObjectNode:
		c.log.Notice("Check object node:")
		c.log.Default(node.IndentedNodeString(0))
		c.checkCompatibilityOfConstraints(node)
		c.checkLinksOfNode(node, ss) // can panic
		c.ensureShortcutKeysAreValid(node)
	case *schema.MixedNode:
		c.log.Notice("Check mixed node:")
		c.log.Default(node.IndentedNodeString(0))
		c.checkCompatibilityOfConstraints(node)
		c.checkLinksOfNode(node, ss) // can panic
	case *schema.MixedValueNode:
		c.log.Notice("Check mixed value node:")
		c.log.Default(node.IndentedNodeString(0))
		c.checkCompatibilityOfConstraints(node)
		c.checkLinksOfNode(node, ss) // can panic
	default:
		panic(errors.ErrImpossible)
	}

	if branchingNode, ok := node.(schema.BranchNode); ok {
		for _, child := range branchingNode.Children() {
			c.checkNode(child, ss) // can panic
		}
	}
}

func (c checkSchema) checkLiteralNode(node schema.Node, ss map[string]schema.Type) {
	checkerList := c.checkerList(node, ss)
	errorsCount := 0
	var err errors.Error

	for _, checker := range checkerList {
		c.log.Notice("\tPut to checker:")
		c.log.Default(checker.indentedString(1))

		err = checker.check(node.BasisLexEventOfSchemaForNode())
		if err == nil {
			c.log.Error("\tSuccess\n")
		} else {
			c.log.Error("\t" + err.Message() + "\n")
			errorsCount++
		}
	}

	if errorsCount == len(checkerList) {
		if len(checkerList) == 1 {
			panic(err)
		} else {
			panic(lexeme.NewLexEventError(node.BasisLexEventOfSchemaForNode(), errors.ErrOrRuleSetValidation))
		}
	}
}

// Checks for array elements. Including recursively for types. Or if the array
// type is "any".
func (c checkSchema) checkArrayItems(node schema.Node) {
	arrayNode := node.(*schema.ArrayNode) //nolint:errcheck // We're sure about this type.

	if arrayNode.Len() != 0 {
		return
	}

	if arrayNode.Constraint(constraint.AnyConstraintType) != nil {
		return
	}

	if typesList := arrayNode.Constraint(constraint.TypesListConstraintType); typesList != nil {
		for _, name := range typesList.(*constraint.TypesList).Names() {
			typeRootNode := c.rootSchema.Type(name).RootNode() // can panic

			if arrayNode, ok := typeRootNode.(*schema.ArrayNode); ok {
				c.checkArrayItems(arrayNode)
			}
		}
	}
}

// check all constraints for compatibility with the json-type of the node
func (checkSchema) checkCompatibilityOfConstraints(node schema.Node) {
	_, isMixed := node.(*schema.MixedNode)
	_, isMixedValue := node.(*schema.MixedValueNode)
	for kv := range node.ConstraintMap().Iterate() {
		if !kv.Value.IsJsonTypeCompatible(node.Type()) && !isMixed && !isMixedValue {
			panic(errors.Format(errors.ErrUnexpectedConstraint, kv.Value.Type().String(), node.RealType()))
		}
	}
}

func (c *checkSchema) checkLinksOfNode(node schema.Node, ss map[string]schema.Type) {
	if node.Constraint(constraint.TypesListConstraintType) == nil {
		return // to optimize memory allocation
	}

	for k := range c.foundTypeNames {
		delete(c.foundTypeNames, k)
	}
	for k := range c.allowedJsonTypes {
		delete(c.allowedJsonTypes, k)
	}

	c.collectAllowedJsonTypes(node, ss)
	if _, ok := c.allowedJsonTypes[node.Type()]; !ok {
		panic(errors.ErrIncorrectUserType)
	}
}

func (c *checkSchema) ensureShortcutKeysAreValid(node *schema.ObjectNode) {
	var lex lexeme.LexEvent

	defer func() {
		r := recover()
		if r == nil {
			return
		}

		err, ok := r.(errors.Errorf)
		if !ok {
			panic(r)
		}

		if err.Code() != errors.ErrTypeNotFound {
			panic(r)
		}

		panic(lexeme.NewLexEventError(lex, err))
	}()

	for kv := range node.Keys().Iterate() {
		if kv.Value.IsShortcut {
			lex = kv.Value.Lex
			c.rootSchema.Type(kv.Key) // can panic
		}
	}
}

func (c *checkSchema) collectAllowedJsonTypes(node schema.Node, ss map[string]schema.Type) {
	if _, ok := node.(*schema.MixedValueNode); ok {
		// This node can be anything.
		for _, t := range json.AllTypes {
			c.allowedJsonTypes[t] = true
		}

		// Check all user types are defined.
		if typesConstraint := node.Constraint(constraint.TypesListConstraintType); typesConstraint != nil {
			for _, typeName := range typesConstraint.(*constraint.TypesList).Names() {
				c.rootSchema.Type(typeName) // can panic
			}
		}
		return
	}

	typesConstraint := node.Constraint(constraint.TypesListConstraintType)

	if typesConstraint == nil {
		c.allowedJsonTypes[node.Type()] = true
		return
	}

	for _, typeName := range typesConstraint.(*constraint.TypesList).Names() {
		if _, ok := c.foundTypeNames[typeName]; ok {
			panic(errors.Format(errors.ErrImpossibleToDetermineTheJsonTypeDueToRecursion, typeName))
		}
		c.foundTypeNames[typeName] = true
		c.collectAllowedJsonTypes(getType(typeName, c.rootSchema, ss).RootNode(), ss) // can panic
	}
}

func getType(n string, rootSchema *schema.Schema, ss map[string]schema.Type) (ret *schema.Schema) {
	getFromRoot := func() *schema.Schema {
		return rootSchema.Type(n)
	}

	getFromMap := func() *schema.Schema {
		s, ok := ss[n]
		if !ok {
			panic(errors.Format(errors.ErrTypeNotFound, n))
		}
		return s.Schema()
	}

	main := getFromRoot
	alternative := getFromMap
	if len(n) > 0 && n[0] == '#' {
		main = getFromMap
		alternative = getFromRoot
	}

	defer func() {
		if r := recover(); r == nil {
			return
		}

		ret = alternative()
	}()
	return main()
}
