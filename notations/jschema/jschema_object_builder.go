package jschema

import (
	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/internal/lexeme"
	"github.com/jsightapi/jsight-schema-go-library/internal/panics"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/loader"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/schema"
)

type ObjectBuilder struct {
	jschema  *Schema
	rootNode *schema.ObjectNode
}

// NewObjectBuilder used only for building Path variables in the JSight API library
func NewObjectBuilder() ObjectBuilder {
	objNode := schema.NewObjectNode(lexeme.LexEvent{})

	inner := schema.New()
	inner.SetRootNode(objNode)

	s := New("", "")
	s.inner = &inner

	return ObjectBuilder{
		jschema:  s,
		rootNode: objNode,
	}
}

func (b ObjectBuilder) AddProperty(key string, node schema.Node, types map[string]schema.Type) {
	k := schema.ObjectNodeKey{
		Key:        key,
		IsShortcut: false,
		Lex:        lexeme.LexEvent{},
	}
	b.rootNode.AddChild(k, node)
	for kk, vv := range types {
		b.jschema.inner.AddType(kk, vv)
	}
}

func (b ObjectBuilder) Len() int {
	return b.rootNode.Len()
}

func (b ObjectBuilder) UserTypeNames() []string {
	b.jschema.collectUserTypes()
	return b.jschema.usedUserTypes.Data()
}

func (b ObjectBuilder) AddType(name string, sc jschema.Schema) error {
	if s, ok := sc.(*Schema); ok {
		b.jschema.inner.AddType(name, schema.Type{
			Schema:   s.inner,
			RootFile: s.file,
		})
	}
	return nil
}

func (b ObjectBuilder) Build() *Schema {
	s := b.jschema
	_ = s.loadOnce.Do(func() (err error) { //nolint:errcheck // It's ok.
		defer func() {
			err = panics.Handle(recover(), err)
		}()
		// sc := loader.LoadSchemaWithoutCompile(
		// 	scanner.New(s.file),
		// 	nil,
		// 	s.rules,
		// )
		// s.inner = &sc
		s.ASTNode = s.buildASTNode()
		// s.collectUserTypes()
		loader.CompileBasic(s.inner, s.areKeysOptionalByDefault)
		return nil
	})

	// _ = s.compileOnce.Do(func() (err error) { //nolint:errcheck // It's ok.
	// 	return nil
	// })

	_ = s.compile() //nolint:errcheck // It's ok.

	return s
}
