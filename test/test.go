package test

import (
	"path/filepath"
	"strings"

	"github.com/jsightapi/jsight-schema-go-library/errors"
)

type test struct {
	relativePath string
	schema       string
	json         string
	types        []string
	enums        []string
	want         errors.Err
}

func (t test) name() string {
	p, err := filepath.Abs(t.String())
	if err != nil {
		panic(err)
	}

	parts := strings.Split(p, string(filepath.Separator))
	var idx int
	for _, p := range parts {
		idx++
		if p == "testdata" {
			break
		}
	}

	return strings.TrimSuffix(filepath.Join(parts[idx:]...), ".json")
}

func (t test) String() string {
	return filepath.Join(t.relativePath, t.json)
}
