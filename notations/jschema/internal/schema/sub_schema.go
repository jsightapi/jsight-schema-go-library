package schema

import (
	"j/schema/bytes"
	"j/schema/fs"
)

type Type struct {
	schema   *Schema
	rootFile *fs.File
	begin    bytes.Index
}

func (s *Type) Schema() *Schema {
	return s.schema
}

func (s *Type) RootFile() *fs.File {
	return s.rootFile
}

func (s *Type) Begin() bytes.Index {
	return s.begin
}
