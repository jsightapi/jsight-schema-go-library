package scanner

import (
	"j/schema/reader"
	"j/schema/test"
	"path/filepath"
	"testing"
)

func BenchmarkSchemaScanner(b *testing.B) {
	file := reader.Read(filepath.Join(test.GetProjectRoot(), "testdata", "big.jschema"))

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		s := NewSchemaScanner(file, false)
		for {
			if _, ok := s.Next(); ok == false {
				break
			}
		}
	}
}
