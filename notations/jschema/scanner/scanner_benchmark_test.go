package scanner

import (
	"path/filepath"
	"testing"

	"github.com/jsightapi/jsight-schema-go-library/reader"
	"github.com/jsightapi/jsight-schema-go-library/test"
)

func BenchmarkScanner(b *testing.B) {
	file := reader.Read(filepath.Join(test.GetProjectRoot(), "testdata", "big.jschema"))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := New(file)
		for {
			if _, ok := s.Next(); ok == false {
				break
			}
		}
	}
}
