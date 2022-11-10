package test

import (
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/formats/json"
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/kit"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema"
	"github.com/jsightapi/jsight-schema-go-library/reader"
	"github.com/jsightapi/jsight-schema-go-library/rules/enum"
)

func TestData(t *testing.T) {
	for _, tt := range tests() {
		t.Run(tt.name(), func(t *testing.T) {
			err := validate(tt)

			if tt.want == nil {
				if err != nil {
					t.Errorf(`Unexpected error
	File: %s
	Position: %d
	tCode: %v
	Message: %s`, err.Filename(), err.Position(), err.ErrCode(), err.Message())
				}
			} else {
				want := (int)(tt.want.Code())
				if err == nil {
					t.Errorf("There must have been a error code: %v", want)
				} else if want != err.ErrCode() {
					t.Errorf(`Invalid error code
	File: %s
	Want error code: %v
	Got error code: %v
	Message: %s`, err.Filename(), want, err.ErrCode(), err.Message())
				}
			}
		})
	}
}

func validate(t test) kit.Error {
	schemaFile := reader.Read(path.Join(GetProjectRoot(), t.relativePath, t.schema))
	jsonFile := reader.Read(path.Join(GetProjectRoot(), t.relativePath, t.json))
	types := readFiles(t.relativePath, t.types)
	enums := readFiles(t.relativePath, t.enums)

	sc := jschema.FromFile(schemaFile)

	for name, f := range enums {
		if f.Content().Len() == 0 {
			return errors.NewDocumentError(schemaFile, errors.Format(errors.ErrEmptyType, name))
		}
		if err := sc.AddRule(name, enum.FromFile(f)); err != nil {
			return kit.ConvertError(f, err)
		}
	}

	for name, f := range types {
		if f.Content().Len() == 0 {
			return errors.NewDocumentError(schemaFile, errors.Format(errors.ErrEmptyType, name))
		}
		if err := sc.AddType(name, jschema.FromFile(f)); err != nil {
			return kit.ConvertError(f, err)
		}
	}

	err := sc.Validate(json.FromFile(jsonFile))
	if err != nil {
		return kit.ConvertError(schemaFile, err)
	}
	return nil
}

func readFiles(relativePath string, filenames []string) map[string]*fs.File {
	types := make(map[string]*fs.File)

	for _, filename := range filenames {
		absolutePath := path.Join(GetProjectRoot(), relativePath, filename)

		ext := filepath.Ext(filename)
		typeName := "@" + strings.TrimSuffix(filename, ext)

		file := reader.ReadWithName(absolutePath, typeName)

		types[typeName] = file
	}

	return types
}
