package ischema

import (
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/fs"
)

type Type struct {
	Schema   *ISchema
	RootFile *fs.File
	Begin    bytes.Index
}
