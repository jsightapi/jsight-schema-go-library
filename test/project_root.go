package test

import (
	"os"
	"path/filepath"

	"github.com/jsightapi/jsight-schema-go-library/internal/sync"
)

var projectRootOnce sync.ErrOnceWithValue[string]

func GetProjectRoot() string {
	v, _ := projectRootOnce.Do(func() (string, error) {
		return determineProjectRoot(), nil
	})
	return v
}

func determineProjectRoot() string {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for {
		if path == "/" || path == "" {
			panic("Project root not found")
		}
		if isExists(filepath.Join(path, "go.mod")) {
			break
		}
		path = filepath.Dir(path)
	}
	return path
}

func isExists(f string) bool {
	_, err := os.Stat(f)
	return err == nil
}
