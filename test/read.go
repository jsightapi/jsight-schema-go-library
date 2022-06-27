package test

import (
	"os"
	"path"
)

func directories() []dir {
	return readTestdataDirectory("testdata")
}

func readTestdataDirectory(relativePath string) []dir {
	absoluteDirPath := path.Join(GetProjectRoot(), relativePath)

	files, err := os.ReadDir(absoluteDirPath)
	if err != nil {
		panic(err)
	}

	directories := make([]dir, 0, 1)
	d := newDir(relativePath)

	for _, file := range files {
		if file.IsDir() {
			child := readTestdataDirectory(path.Join(relativePath, file.Name()))
			directories = append(directories, child...)
		} else {
			d.appendFilename(file.Name())
		}
	}

	if !d.isEmpty() {
		directories = append(directories, d)
	}

	return directories
}
