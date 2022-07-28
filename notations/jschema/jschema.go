package jschema

import (
	stdBytes "bytes"
	stdErrors "errors"
	"fmt"
	"io"
	"strings"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/formats/json"
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/internal/panics"
	"github.com/jsightapi/jsight-schema-go-library/internal/sync"
	"github.com/jsightapi/jsight-schema-go-library/notations/internal"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/checker"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/loader"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/scanner"
	internalSchema "github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema/constraint"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/validator"
	"github.com/jsightapi/jsight-schema-go-library/notations/regex"
)

type Schema struct {
	file  *fs.File
	inner *internalSchema.Schema

	rules map[string]jschema.Rule

	usedUserTypes []string

	lenOnce     sync.ErrOnceWithValue[uint]
	loadOnce    sync.ErrOnce
	compileOnce sync.ErrOnce

	astNode                  jschema.ASTNode
	areKeysOptionalByDefault bool
}

var _ jschema.Schema = &Schema{}

// New creates a Jsight schema with specified name and content.
func New[T fs.FileContent](name string, content T, oo ...Option) *Schema {
	return FromFile(fs.NewFile(name, content), oo...)
}

// FromFile creates a Jsight schema from file.
func FromFile(f *fs.File, oo ...Option) *Schema {
	s := &Schema{
		file:  f,
		rules: map[string]jschema.Rule{},
	}

	for _, o := range oo {
		o(s)
	}

	return s
}

type Option func(s *Schema)

func KeysAreOptionalByDefault() Option {
	return func(s *Schema) {
		s.areKeysOptionalByDefault = true
	}
}

func (s *Schema) Len() (uint, error) {
	return s.lenOnce.Do(func() (uint, error) {
		return s.computeLen()
	})
}

func (s *Schema) computeLen() (length uint, err error) {
	// Iterate through all lexemes until we reach the end
	// We should rewind here in case we call NextLexeme method.
	defer func() {
		err = panics.Handle(recover(), err)
	}()

	return scanner.New(s.file, scanner.ComputeLength).Length(), err
}

func (s *Schema) Example() (b []byte, err error) {
	defer func() {
		err = panics.Handle(recover(), err)
	}()

	if err := s.compile(); err != nil {
		return nil, err
	}

	if s.inner.RootNode() == nil {
		return nil, errors.NewDocumentError(s.file, errors.ErrEmptySchema)
	}

	return newExampleBuilder(s.inner.TypesList()).Build(s.inner.RootNode())
}

type exampleBuilder struct {
	// types all user types used in this schema.
	types map[string]internalSchema.Type

	// processedTypes an unordered set of processed types required for handling
	// recursion.
	// Infinity recursion can't happen here 'cause we check it before building
	// example, but optional recursion can be there.
	processedTypes map[string]int
}

func newExampleBuilder(types map[string]internalSchema.Type) *exampleBuilder {
	return &exampleBuilder{
		types:          types,
		processedTypes: map[string]int{},
	}
}

func (b *exampleBuilder) Build(node internalSchema.Node) ([]byte, error) {
	switch typedNode := node.(type) {
	case *internalSchema.ObjectNode:
		return b.buildExampleForObjectNode(typedNode)

	case *internalSchema.ArrayNode:
		return b.buildExampleForArrayNode(typedNode)

	case *internalSchema.LiteralNode:
		return typedNode.BasisLexEventOfSchemaForNode().Value(), nil

	case *internalSchema.MixedValueNode:
		return b.buildExampleForMixedValueNode(typedNode)

	default:
		return nil, fmt.Errorf("unhandled node type %T", node)
	}
}

func (b *exampleBuilder) buildExampleForObjectNode(node *internalSchema.ObjectNode) ([]byte, error) {
	if node.Constraint(constraint.TypesListConstraintType) != nil {
		return nil, errors.ErrUserTypeFound
	}

	buf := exampleBufferPool.Get()
	defer exampleBufferPool.Put(buf)

	buf.WriteRune('{')
	children := node.Children()
	length := len(children)
	for i, childNode := range children {
		ex, err := b.Build(childNode)
		if err != nil {
			return nil, err
		}

		if ex == nil {
			continue
		}

		k, err := b.buildObjectKey(node.Key(i))
		if err != nil {
			return nil, err
		}

		buf.WriteRune('"')
		buf.Write(k)
		buf.WriteString(`":`)
		buf.Write(ex)
		if i+1 != length {
			buf.WriteRune(',')
		}
	}
	buf.WriteRune('}')
	return buf.Bytes(), nil
}

func (b *exampleBuilder) buildObjectKey(k internalSchema.ObjectNodeKey) ([]byte, error) {
	if !bytes.Bytes(k.Key).IsUserTypeName() {
		return []byte(k.Key), nil
	}

	typ, ok := b.types[k.Key]
	if !ok {
		return nil, errors.Format(errors.ErrUnknownType, k.Key)
	}

	ex, err := b.Build(typ.Schema().RootNode())
	if err != nil {
		return nil, err
	}
	return stdBytes.Trim(ex, `"`), nil
}

func (b *exampleBuilder) buildExampleForArrayNode(node *internalSchema.ArrayNode) ([]byte, error) {
	if node.Constraint(constraint.TypesListConstraintType) != nil {
		return nil, errors.ErrUserTypeFound
	}

	buf := exampleBufferPool.Get()
	defer exampleBufferPool.Put(buf)

	buf.WriteRune('[')
	children := node.Children()
	length := len(children)
	for i, childNode := range children {
		ex, err := b.Build(childNode)
		if err != nil {
			return nil, err
		}

		if ex == nil {
			continue
		}

		buf.Write(ex)
		if i+1 != length {
			buf.WriteRune(',')
		}
	}
	buf.WriteRune(']')
	return buf.Bytes(), nil
}

func (b *exampleBuilder) buildExampleForMixedValueNode(node *internalSchema.MixedValueNode) ([]byte, error) {
	tt := node.GetTypes()
	if len(tt) == 0 {
		// Normally this shouldn't happen, but we still have to handle this case.
		return nil, errors.ErrLoader
	}

	typeName := tt[0]
	if !bytes.Bytes(typeName).IsUserTypeName() {
		return node.Value(), nil
	}

	if cnt := b.processedTypes[typeName]; cnt > 1 {
		// Do not process already processed type more than twice.
		return nil, nil
	}

	b.processedTypes[typeName]++
	defer func() {
		delete(b.processedTypes, typeName)
	}()

	t, ok := b.types[typeName]
	if !ok {
		return nil, errors.Format(errors.ErrTypeNotFound, typeName)
	}
	return b.Build(t.Schema().RootNode())
}

func buildExample(node internalSchema.Node, types map[string]internalSchema.Type) ([]byte, error) {
	switch typedNode := node.(type) {
	case *internalSchema.ObjectNode:
		return buildExampleForObjectNode(typedNode, types)

	case *internalSchema.ArrayNode:
		return buildExampleForArrayNode(typedNode, types)

	case *internalSchema.LiteralNode:
		return typedNode.BasisLexEventOfSchemaForNode().Value(), nil

	case *internalSchema.MixedValueNode:
		return buildExampleForMixedValueNode(typedNode, types)

	default:
		return nil, fmt.Errorf("unhandled node type %T", node)
	}
}

func buildExampleForObjectNode(
	node *internalSchema.ObjectNode,
	types map[string]internalSchema.Type,
) ([]byte, error) {
	if node.Constraint(constraint.TypesListConstraintType) != nil {
		return nil, errors.ErrUserTypeFound
	}

	b := exampleBufferPool.Get()
	defer exampleBufferPool.Put(b)

	b.WriteRune('{')
	children := node.Children()
	length := len(children)
	for i, childNode := range children {
		key := node.Key(i)
		b.WriteRune('"')
		b.WriteString(key.Key)
		b.WriteString(`":`)

		ex, err := buildExample(childNode, types)
		if err != nil {
			return nil, err
		}
		b.Write(ex)
		if i+1 != length {
			b.WriteRune(',')
		}
	}
	b.WriteRune('}')
	return b.Bytes(), nil
}

func buildExampleForArrayNode(
	node *internalSchema.ArrayNode,
	types map[string]internalSchema.Type,
) ([]byte, error) {
	if node.Constraint(constraint.TypesListConstraintType) != nil {
		return nil, errors.ErrUserTypeFound
	}

	b := exampleBufferPool.Get()
	defer exampleBufferPool.Put(b)

	b.WriteRune('[')
	children := node.Children()
	length := len(children)
	for i, childNode := range children {
		ex, err := buildExample(childNode, types)
		if err != nil {
			return nil, err
		}
		b.Write(ex)
		if i+1 != length {
			b.WriteRune(',')
		}
	}
	b.WriteRune(']')
	return b.Bytes(), nil
}

var exampleBufferPool = sync.NewBufferPool(512)

func buildExampleForMixedValueNode(
	node *internalSchema.MixedValueNode,
	types map[string]internalSchema.Type,
) ([]byte, error) {
	tt := node.GetTypes()
	if len(tt) == 0 {
		// Normally this shouldn't happen, but we still have to handle this case.
		return nil, errors.ErrLoader
	}

	typeName := tt[0]
	if !bytes.Bytes(typeName).IsUserTypeName() {
		return node.Value(), nil
	}

	t, ok := types[typeName]
	if !ok {
		return nil, errors.Format(errors.ErrTypeNotFound, typeName)
	}
	return buildExample(t.Schema().RootNode(), types)
}

func (s *Schema) AddType(name string, sc jschema.Schema) (err error) {
	defer func() {
		err = panics.Handle(recover(), err)
	}()

	if err := s.load(); err != nil {
		return err
	}

	switch typ := sc.(type) {
	case *Schema:
		if err := typ.load(); err != nil {
			return fmt.Errorf("load added type: %w", err)
		}

		s.inner.AddNamedType(name, typ.inner, s.file, 0)
	case *regex.Schema:
		pattern, err := typ.Pattern()
		if err != nil {
			return err
		}

		example, err := typ.Example()
		if err != nil {
			return fmt.Errorf("generate example for Regex type: %w", err)
		}

		typSc := New(name, fmt.Sprintf("%q // {regex: %q}", example, pattern))
		if err := typSc.load(); err != nil {
			return fmt.Errorf("load added type: %w", err)
		}

		s.inner.AddNamedType(name, typSc.inner, s.file, 0)

	default:
		return fmt.Errorf("schema should be JSight or Regex schema, but %T given", sc)
	}

	return nil
}

func (s *Schema) AddRule(n string, r jschema.Rule) error {
	if s.inner != nil {
		return stdErrors.New("schema is already compiled")
	}

	if r == nil {
		return stdErrors.New("rule is nil")
	}

	if err := r.Check(); err != nil {
		return err
	}
	s.rules[n] = r
	return nil
}

func (s *Schema) Check() (err error) {
	defer func() {
		err = panics.Handle(recover(), err)
	}()
	return s.compile()
}

func (s *Schema) Validate(document jschema.Document) (err error) {
	defer func() {
		err = panics.Handle(recover(), err)
	}()
	if err := s.compile(); err != nil {
		return err
	}

	if _, ok := document.(*json.Document); !ok {
		return fmt.Errorf("support only JSON documents, but got %T", document)
	}

	return s.validate(document)
}

func (s *Schema) validate(document jschema.Document) error {
	tree := validator.NewTree(
		validator.NodeValidatorList(s.inner.RootNode(), *s.inner, nil),
	)

	empty := true

	for {
		jsonLex, err := document.NextLexeme()
		if err != nil {
			if stdErrors.Is(err, io.EOF) {
				break
			}
			return err
		}

		empty = false
		if tree.FeedLeaves(jsonLex) { // can panic: error of validation
			break
		}
	}

	if empty {
		return internal.NewValidatorError(errors.ErrEmptyJson, "")
	}

	// check for error: Invalid non-space byte after top-level value
	for {
		_, err := document.NextLexeme()
		if err != nil {
			if stdErrors.Is(err, io.EOF) {
				break
			}
			return err
		}
	}
	return nil
}

func (s *Schema) GetAST() (an jschema.ASTNode, err error) {
	if err := s.compile(); err != nil {
		return jschema.ASTNode{}, err
	}

	return s.astNode, nil
}

func (s *Schema) UsedUserTypes() ([]string, error) {
	if err := s.load(); err != nil {
		return nil, err
	}
	return s.usedUserTypes, nil
}

func (s *Schema) load() error {
	return s.loadOnce.Do(func() (err error) {
		defer func() {
			err = panics.Handle(recover(), err)
		}()
		sc := loader.LoadSchemaWithoutCompile(
			scanner.New(s.file),
			nil,
			s.rules,
		)
		s.inner = &sc
		s.astNode = s.buildASTNode()
		s.collectUserTypes()
		loader.CompileBasic(s.inner, s.areKeysOptionalByDefault)
		return nil
	})
}

func (s *Schema) Build() error {
	return s.compile()
}

func (s *Schema) collectUserTypes() {
	node := s.inner.RootNode()
	// This is possible when schema isn't valid.
	if node == nil {
		return
	}

	s.usedUserTypes = collectUserTypes(node)
}

func collectUserTypes(node internalSchema.Node) []string {
	c := &userTypesCollector{
		alreadyProcessed: map[string]struct{}{},
	}
	c.collect(node)
	return c.userTypes
}

type userTypesCollector struct {
	alreadyProcessed map[string]struct{}
	userTypes        []string
}

func (c *userTypesCollector) collect(node internalSchema.Node) {
	c.collectUserTypesFromTypesListConstraint(node)
	c.collectUserTypesFromTypeConstraint(node)
	c.collectUserTypesFromAllOfConstraint(node)

	switch n := node.(type) {
	case *internalSchema.ObjectNode:
		c.collectUserTypesObjectNode(n)

	case *internalSchema.ArrayNode:
		for _, child := range n.Children() {
			c.collect(child)
		}

	case *internalSchema.MixedValueNode:
		for _, ut := range strings.Split(n.Value().String(), "|") {
			s := strings.TrimSpace(ut)
			if s[0] == '@' {
				c.addType(s)
			}
		}
	}
}

func (c *userTypesCollector) collectUserTypesFromTypesListConstraint(node internalSchema.Node) {
	cnstr := node.Constraint(constraint.TypesListConstraintType)
	if cnstr == nil {
		return
	}

	list, ok := cnstr.(*constraint.TypesList)
	if !ok {
		return
	}

	for _, name := range list.Names() {
		if name[0] == '@' {
			c.addType(name)
		}
	}
}

func (c *userTypesCollector) collectUserTypesFromTypeConstraint(node internalSchema.Node) {
	cnstr := node.Constraint(constraint.TypeConstraintType)
	if cnstr == nil {
		return
	}

	typ, ok := cnstr.(*constraint.TypeConstraint)
	if !ok {
		return
	}

	name := typ.Bytes().Unquote().String()
	if name[0] == '@' {
		c.addType(name)
	}
}

func (c *userTypesCollector) collectUserTypesFromAllOfConstraint(node internalSchema.Node) {
	cnstr := node.Constraint(constraint.AllOfConstraintType)
	if c == nil {
		return
	}

	allOf, ok := cnstr.(*constraint.AllOf)
	if !ok {
		return
	}

	for _, name := range allOf.SchemaNames() {
		if name[0] == '@' {
			c.addType(name)
		}
	}
}

func (c *userTypesCollector) collectUserTypesObjectNode(node *internalSchema.ObjectNode) {
	for _, v := range node.Keys().Data {
		k := v.Key

		if v.IsShortcut {
			if k[0] == '@' {
				c.addType(k)
			}
		}

		child, ok := node.Child(k)
		if ok {
			c.collect(child)
		}
	}
}

func (c *userTypesCollector) addType(n string) {
	if _, ok := c.alreadyProcessed[n]; ok {
		return
	}
	c.alreadyProcessed[n] = struct{}{}
	c.userTypes = append(c.userTypes, n)
}

func (s *Schema) buildASTNode() jschema.ASTNode {
	root := s.inner.RootNode()
	if root == nil {
		// This case will be handled in loader.CompileBasic.
		return jschema.ASTNode{
			Rules: &jschema.RuleASTNodes{},
		}
	}

	an, err := root.ASTNode()
	if err != nil {
		panic(err)
	}
	return an
}

func (s *Schema) compile() error {
	return s.compileOnce.Do(func() (err error) {
		defer func() {
			err = panics.Handle(recover(), err)
		}()
		if err := s.load(); err != nil {
			return err
		}
		loader.CompileAllOf(s.inner)
		loader.AddUnnamedTypes(s.inner)
		checker.CheckRootSchema(s.inner)
		return checker.CheckRecursion(s.file.Name(), s.inner)
	})
}
