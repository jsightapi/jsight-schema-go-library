package test

import (
	"j/schema/errors"
	"path/filepath"
	"strings"
)

type test struct {
	relativePath string
	schema       string
	json         string
	types        []string
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
