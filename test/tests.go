package test

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/jsightapi/jsight-schema-go-library/errors"
)

func tests() []test {
	tests := make([]test, 0, 100)
	for _, d := range directories() {
		tests = append(tests, dirTests(d)...)
	}
	return tests
}

func dirTests(d dir) []test {
	list := make([]test, 0, 10)
	for _, jsonFilename := range d.json {
		t := test{
			relativePath: d.relativePath,
			schema:       d.schema,
			json:         jsonFilename,
			types:        d.types,
			enums:        d.enums,
			want:         want(jsonFilename),
		}
		list = append(list, t)
	}
	return list
}

// want determines the expected error code by the file name.
// If the file starts with "err_", then an error code is expected further in the file name.
// For example, if the file name is "err_801_something_else.json", the error code will be 801.
// If the file name is "some_name.json", then the error code will be nil.
func want(filename string) errors.Err {
	ext := filepath.Ext(filename)
	p := strings.Split(strings.TrimSuffix(filename, ext), "_")

	if p[0] == "err" {
		code, err := strconv.Atoi(p[1])
		if err != nil {
			panic("Invalid error code in the file name: " + p[1])
		}
		return errors.ErrorCode(code)
	}

	return nil
}
