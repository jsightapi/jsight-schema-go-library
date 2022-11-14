package checker

import (
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/json"
	"github.com/jsightapi/jsight-schema-go-library/lexeme"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/ischema"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/ischema/constraint"
)

// Checks the SAMPLE SCHEMA and all TYPES for compliance with all RULES.

type checkSchema struct {
	rootSchema *ischema.ISchema

	// foundTypeNames the names of the type encountered during checking. Are used
	// to control recursion.
	foundTypeNames map[string]struct{}

	// allowedJsonTypes the list of available json-types from types.
	allowedJsonTypes map[json.Type]struct{}
}

func CheckRootSchema(rootSchema *ischema.ISchema) {
	c := checkSchema{
		rootSchema:       rootSchema,
		foundTypeNames:   make(map[string]struct{}, 10),
		allowedJsonTypes: make(map[json.Type]struct{}, 10),
	}

	if rootSchema.RootNode() != nil { // the root schema may contain no nodes
		c.checkNode(rootSchema.RootNode(), rootSchema.TypesList())
	}

	for name, typ := range rootSchema.TypesList() {
		c.checkType(name, typ, rootSchema.TypesList())
	}
}

func (c *checkSchema) checkType(name string, typ ischema.Type, ss map[string]ischema.Type) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}

		// Return an error with the full set of bytes of the root schema.
		if documentError, ok := r.(errors.DocumentError); ok {
			documentError.SetFile(typ.RootFile)
			documentError.SetIndex(documentError.Index() + typ.Begin)
			documentError.SetIncorrectUserType(name)
			panic(documentError)
		}

		panic(r)
	}()

	c.checkNode(typ.Schema.RootNode(), ss)
}

func (c *checkSchema) checkerList(node ischema.Node, ss map[string]ischema.Type) []nodeChecker {
	l := nodeCheckerListConstructor{
		rootSchema: c.rootSchema,
		types:      ss,
	}
	l.buildList(node)
	return l.list
}

func (c checkSchema) checkNode(node ischema.Node, ss map[string]ischema.Type) {
	defer lexeme.CatchLexEventError(node.BasisLexEventOfSchemaForNode())
	switch node := node.(type) {
	case *ischema.LiteralNode:
		c.checkCompatibilityOfConstraints(node)
		c.checkLinksOfNode(node, ss) // can panic
		c.checkLiteralNode(node, ss)
	case *ischema.ArrayNode:
		c.checkCompatibilityOfConstraints(node)
		c.checkLinksOfNode(node, ss) // can panic
		c.checkArrayItems(node)
		c.checkArrayNode(node)
	case *ischema.ObjectNode:
		c.checkCompatibilityOfConstraints(node)
		c.checkLinksOfNode(node, ss) // can panic
		if err := c.ensureShortcutKeysAreValid(node); err != nil {
			panic(err)
		}
		c.checkAdditionalPropertiesConstraint(node, ss)
	case *ischema.MixedNode:
		c.checkCompatibilityOfConstraints(node)
		c.checkLinksOfNode(node, ss) // can panic
	case *ischema.MixedValueNode:
		c.checkCompatibilityOfConstraints(node)
		c.checkLinksOfNode(node, ss) // can panic
	default:
		panic(errors.ErrImpossible)
	}

	if branchingNode, ok := node.(ischema.BranchNode); ok {
		for _, child := range branchingNode.Children() {
			c.checkNode(child, ss) // can panic
		}
	}
}

func (c checkSchema) checkLiteralNode(node ischema.Node, ss map[string]ischema.Type) {
	checkerList := c.checkerList(node, ss)
	errorsCount := 0
	var err errors.Error

	for _, checker := range checkerList {
		err = checker.Check(node.BasisLexEventOfSchemaForNode())
		if err != nil {
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
func (c checkSchema) checkArrayItems(node ischema.Node) {
	arrayNode := node.(*ischema.ArrayNode)

	if arrayNode.Len() != 0 {
		return
	}

	if arrayNode.Constraint(constraint.AnyConstraintType) != nil {
		return
	}

	if typesList := arrayNode.Constraint(constraint.TypesListConstraintType); typesList != nil {
		for _, name := range typesList.(*constraint.TypesList).Names() {
			typeRootNode := c.rootSchema.MustType(name).RootNode() // can panic

			if arrayNode, ok := typeRootNode.(*ischema.ArrayNode); ok {
				c.checkArrayItems(arrayNode)
			}
		}
	}
}

func (checkSchema) checkArrayNode(node ischema.Node) {
	arrayNode := node.(*ischema.ArrayNode)

	length := uint(arrayNode.Len())

	if cnstr := arrayNode.Constraint(constraint.MinItemsConstraintType); cnstr != nil {
		cnstr.(*constraint.MinItems).ValidateTheArray(length)
	}

	if cnstr := arrayNode.Constraint(constraint.MaxItemsConstraintType); cnstr != nil {
		cnstr.(*constraint.MaxItems).ValidateTheArray(length)
	}
}

// check all constraints for compatibility with the json-type of the node
func (checkSchema) checkCompatibilityOfConstraints(node ischema.Node) {
	_, isMixed := node.(*ischema.MixedNode)
	_, isMixedValue := node.(*ischema.MixedValueNode)

	err := node.ConstraintMap().Each(func(k constraint.Type, v constraint.Constraint) error {
		if !v.IsJsonTypeCompatible(node.Type()) && !isMixed && !isMixedValue {
			return errors.Format(errors.ErrUnexpectedConstraint, v.Type().String(), node.RealType())
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func (c *checkSchema) checkLinksOfNode(node ischema.Node, ss map[string]ischema.Type) {
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

func (c *checkSchema) ensureShortcutKeysAreValid(node *ischema.ObjectNode) error {
	for _, v := range node.Keys().Data {
		if !v.IsShortcut {
			continue
		}

		s, err := c.rootSchema.Type(v.Key)
		if err != nil {
			return lexeme.NewLexEventError(v.Lex, err)
		}
		actualType := actualRootType(s, c.rootSchema)

		if actualType != json.TypeString {
			return lexeme.NewLexEventError(
				v.Lex,
				errors.Format(errors.ErrInvalidKeyShortcutType, v.Key, actualType),
			)
		}
	}
	return nil
}

func actualRootType(s, root *ischema.ISchema) json.Type {
	t := s.RootNode().Type()
	if t != json.TypeMixed {
		return t
	}

	// mixed type for example: @aaa | @bbb
	if n, ok := s.RootNode().(*ischema.MixedValueNode); ok {
		types := make(map[json.Type]struct{}, 2)
		var tt json.Type
		for _, tn := range n.GetTypes() {
			ss, err := root.Type(tn)
			if err != nil {
				return json.TypeMixed
			}
			tt = actualRootType(ss, root)
			types[tt] = struct{}{}
		}
		if len(types) == 1 { // all USER TYPES (example: @aaa | @bbb) have the same type (example: string)
			return tt
		}
	}

	return json.TypeMixed
}

func (c *checkSchema) collectAllowedJsonTypes(node ischema.Node, ss map[string]ischema.Type) {
	if _, ok := node.(*ischema.MixedValueNode); ok {
		// This node can be anything.
		for _, t := range json.AllTypes {
			c.allowedJsonTypes[t] = struct{}{}
		}

		// Check all user types are defined.
		if typesConstraint := node.Constraint(constraint.TypesListConstraintType); typesConstraint != nil {
			for _, typeName := range typesConstraint.(*constraint.TypesList).Names() {
				c.rootSchema.MustType(typeName) // can panic
			}
		}
		return
	}

	typesConstraint := node.Constraint(constraint.TypesListConstraintType)

	if typesConstraint == nil {
		c.allowedJsonTypes[node.Type()] = struct{}{}
		return
	}

	for _, typeName := range typesConstraint.(*constraint.TypesList).Names() {
		if _, ok := c.foundTypeNames[typeName]; ok {
			panic(errors.Format(errors.ErrImpossibleToDetermineTheJsonTypeDueToRecursion, typeName))
		}
		c.foundTypeNames[typeName] = struct{}{}
		c.collectAllowedJsonTypes(getType(typeName, c.rootSchema, ss).RootNode(), ss) // can panic
	}
}

func (c *checkSchema) checkAdditionalPropertiesConstraint(node ischema.Node, ss map[string]ischema.Type) {
	cnstr := node.Constraint(constraint.AdditionalPropertiesConstraintType)
	if c == nil {
		return
	}

	ap, ok := cnstr.(*constraint.AdditionalProperties)
	if !ok {
		return
	}

	if ap.Mode() == constraint.AdditionalPropertiesMustBeUserType {
		getType(ap.TypeName().String(), c.rootSchema, ss)
	}
}

func getType(n string, rootSchema *ischema.ISchema, ss map[string]ischema.Type) (ret *ischema.ISchema) {
	getFromRoot := func() *ischema.ISchema {
		return rootSchema.MustType(n)
	}

	getFromMap := func() *ischema.ISchema {
		s, ok := ss[n]
		if !ok {
			panic(errors.Format(errors.ErrTypeNotFound, n))
		}
		return s.Schema
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
