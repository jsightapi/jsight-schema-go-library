package schema

import (
	"fmt"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/fs"
)

type Schema struct {
	// types the map where key is the name of the type (or included Schema).
	types    map[string]Type
	rootNode Node
}

func New() Schema {
	return Schema{
		types: make(map[string]Type, 5),
	}
}

func (s Schema) TypesList() map[string]Type {
	return s.types
}

// MustType returns *Schema or panic if not found.
// Deprecated: use Schema.MustType instead
func (s Schema) MustType(name string) *Schema {
	t, ok := s.types[name]
	if ok {
		return t.schema
	}
	panic(errors.Format(errors.ErrTypeNotFound, name))
}

// Type returns specified type's schema.
func (s Schema) Type(name string) (*Schema, errors.Err) {
	t, ok := s.types[name]
	if ok {
		return t.schema, nil
	}
	return nil, errors.Format(errors.ErrTypeNotFound, name)
}

func (s Schema) RootNode() Node {
	return s.rootNode
}

func (s *Schema) AddNamedType(name string, typ *Schema, rootFile *fs.File, begin bytes.Index) {
	if !bytes.Bytes(name).IsUserTypeName() {
		panic(errors.Format(errors.ErrInvalidSchemaName, name))
	}
	s.addType(name, typ, rootFile, begin)
}

// AddUnnamedType Adds an unnamed TYPE to the SCHEMA. Returns a unique name for the added TYPE.
func (s *Schema) AddUnnamedType(typ *Schema, rootFile *fs.File, begin bytes.Index) string {
	name := fmt.Sprintf("#%p", typ)
	s.addType(name, typ, rootFile, begin)
	return name
}

func (s *Schema) addType(name string, schema *Schema, rootFile *fs.File, begin bytes.Index) {
	if _, ok := s.types[name]; ok {
		panic(errors.Format(errors.ErrDuplicationOfNameOfTypes, name))
	}
	s.types[name] = Type{schema, rootFile, begin}
}

func (s *Schema) AddType(n string, t Type) {
	s.types[n] = t
}

func (s *Schema) SetRootNode(node Node) {
	s.rootNode = node
}
