package jschema

import (
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	jerr "github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
	"github.com/jsightapi/jsight-schema-go-library/internal/panics"
	"github.com/jsightapi/jsight-schema-go-library/kit"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/loader"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/scanner"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/validator"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/schema"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/schema/constraint"
)

// NewRawPathVariablesSchema used in jsight api go library
func NewRawPathVariablesSchema(content bytes.Bytes, userTypes map[string]*Schema) (*Schema, error) {
	s := New("", content)

	err := s.loadPathVariables()
	if err != nil {
		return nil, err
	}

	for k, v := range userTypes {
		if err = s.AddType(k, v); err != nil {
			return nil, err
		}
	}

	err = s.compilePathVariables()
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Schema) loadPathVariables() error {
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
		s.ASTNode = s.buildASTNode()
		// s.collectUserTypes()
		// loader.CompileBasic(s.inner, s.areKeysOptionalByDefault)
		return nil
	})
}

func (s *Schema) compilePathVariables() error {
	return s.compileOnce.Do(func() (err error) {
		defer func() {
			err = panics.Handle(recover(), err)
		}()
		loader.CompileAllOf(s.inner)
		// loader.AddUnnamedTypes(s.inner)
		// checker.CheckRootSchema(s.inner)
		// return checker.CheckRecursion(s.file.Name(), s.inner)
		return nil
	})
}

func (s *Schema) RootObjectNode() (*schema.ObjectNode, bool) {
	root := s.inner.RootNode()
	if root.Type() != json.TypeObject {
		// return nil, kit.ConvertError(s.file, jerr.ErrObjectExpected)
		return nil, false
	}

	obj, ok := root.(*schema.ObjectNode)
	if !ok {
		// return nil, kit.ConvertError(s.file, jerr.ErrObjectExpected)
		return nil, false
	}

	return obj, true
}

// ObjectFirstLevelProperties used to collect Path variables in the JSight API library
func (s *Schema) ObjectFirstLevelProperties(ut map[string]*Schema) map[string]schema.Node {
	m := make(map[string]schema.Node, 5)
	s.objectFirstLevelProperties(m, ut)
	return m
}

func (s *Schema) objectFirstLevelProperties(m map[string]schema.Node, ut map[string]*Schema) {
	s.appendPropertiesFromShortcut(m, ut)
	s.appendPropertiesFromObject(m)
}

func (s *Schema) appendPropertiesFromShortcut(m map[string]schema.Node, ut map[string]*Schema) {
	c := s.inner.RootNode().Constraint(constraint.TypeConstraintType)
	if c == nil {
		return
	}

	t, ok := c.(*constraint.TypeConstraint)
	if !ok {
		return
	}

	if ss, ok := ut[t.Bytes().String()]; ok {
		ss.objectFirstLevelProperties(m, ut)
	}
}

func (s *Schema) appendPropertiesFromObject(m map[string]schema.Node) {
	obj, ok := s.RootObjectNode()
	if !ok {
		return
	}

	for _, v := range obj.Keys().Data {
		if !v.IsShortcut {
			if n, ok := obj.Child(v.Key, false); ok {
				n.SetParent(nil)
				m[v.Key] = n
			}
		}
	}
}

func (s *Schema) ObjectProperty(key string) (schema.Node, bool) {
	obj, ok := s.RootObjectNode()
	if !ok {
		return nil, false
	}
	return obj.Child(key, false)
}

func (s *Schema) ValidateObjectProperty(key, value string) (err kit.Error) {
	defer func() {
		if r := recover(); r != nil {
			err = kit.ConvertError(s.file, r)
		}
	}()

	node, ok := s.ObjectProperty(key)
	if !ok {
		return kit.ConvertError(s.file, jerr.Format(jerr.ErrPropertyNotFound, key))
	}

	return validator.ValidateLiteralValue2(node, *s.inner, bytes.NewBytes(value))
}
