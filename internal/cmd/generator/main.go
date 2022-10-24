// This generator should be used for generating some source code.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	if err := run(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func run() error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	g := collectionOfGenerators{
		gg: []generator{
			orderedMapGenerator{},
			stringerGenerator{},
			newSetGenerator(),
		},
	}

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		name := info.Name()
		if info.IsDir() {
			if shouldIgnoreDir(name) {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldIgnoreFile(name) {
			return nil
		}

		log.Printf("Process file %q", path)
		return g.Generate(path)
	})
}

func shouldIgnoreDir(name string) bool {
	return (name == "vendor") ||
		(name[0] == '.') ||
		(name == "cmd") ||
		(name == "test") ||
		(name == "testdata")
}

func shouldIgnoreFile(name string) bool {
	return !strings.HasSuffix(name, ".go") ||
		strings.HasSuffix(name, "_test.go") ||
		strings.HasSuffix(name, "_gen.go")
}

type generator interface {
	Generate(string) error
}

type collectionOfGenerators struct {
	gg []generator
}

func (c collectionOfGenerators) Generate(p string) error {
	for _, g := range c.gg {
		if err := g.Generate(p); err != nil {
			return err
		}
	}
	return nil
}
