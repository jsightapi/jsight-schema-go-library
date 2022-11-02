package schema

import (
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/fs"
)

type Type struct {
	Schema   *Schema
	RootFile *fs.File
	Begin    bytes.Index
}
