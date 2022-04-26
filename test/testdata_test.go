package test

import (
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/kit"
	"github.com/jsightapi/jsight-schema-go-library/reader"
)

func TestData(t *testing.T) {
	for _, tt := range tests() {
		t.Run(tt.name(), func(t *testing.T) {
			err := validate(tt)

			if tt.want == nil {
				if err != nil {
					t.Errorf("Unexpected error\n\tFile: %s\n\tCode: %v\n\tMessage: %s", err.Filename(), err.ErrCode(), err.Message())
				}
			} else {
				want := (int)(tt.want.Code())
				if err == nil {
					t.Errorf("There must have been a error code: %v", want)
				} else if want != err.ErrCode() {
					t.Errorf("Invalid error code\n\tFile: %s\n\tWant error code: %v\n\tGot error code: %v\n\tMessage: %s", err.Filename(), want, err.ErrCode(), err.Message())
				}
			}
		})
	}
}

func validate(t test) kit.Error {
	schemaFile := reader.Read(path.Join(GetProjectRoot(), t.relativePath, t.schema))
	jsonFile := reader.Read(path.Join(GetProjectRoot(), t.relativePath, t.json))
	types := readTypes(t.relativePath, t.types)

	err := kit.ValidateJson(schemaFile, types, jsonFile, false)

	return err
}

func readTypes(relativePath string, filenames []string) map[string]*fs.File {
	types := make(map[string]*fs.File)

	for _, filename := range filenames {
		absolutePath := path.Join(GetProjectRoot(), relativePath, filename)

		ext := filepath.Ext(filename)
		typeName := "@" + strings.TrimSuffix(filename, ext)

		file := reader.Read(absolutePath)
		file.SetName(typeName)

		types[typeName] = file
	}

	return types
}
