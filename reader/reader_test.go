package reader

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/test"
)

func TestReadFile(t *testing.T) {
	filename := filepath.Join(test.GetProjectRoot(), "testdata", "examples", "boolean", "boolean.jschema")
	content := bytes.Bytes(`true // Schema containing a literal example`)

	file := fs.NewFile(filename, content)

	if !reflect.DeepEqual(file, Read(filename)) {
		t.Error("Incorrect return content")
	}
}

func TestReadFilePanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Panic was expected")
		}
	}()

	Read("not_existing_file.ext")
}
