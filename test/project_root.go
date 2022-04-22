package test

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	projectRootOnce sync.Once
	projectRoot     string
)

func GetProjectRoot() string {
	projectRootOnce.Do(func() {
		projectRoot = determineProjectRoot()
	})
	return projectRoot
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

// Integer power: compute a**b using binary powering algorithm
// See Donald Knuth, The Art of Computer Programming, Volume 2, Section
// func Pow(a, b uint) uint {
// 	var p uint = 1
// 	for b > 0 {
// 		if b&1 != 0 {
// 			p *= a
// 		}
// 		b >>= 1
// 		a *= a
// 	}
// 	return p
// }
//
// func VarDump(mixed interface{}) string {
// 	b, err := json.MarshalIndent(mixed, "", "  ")
// 	if err != nil {
// 		panic(err)
// 	}
// 	return string(b)
// }
