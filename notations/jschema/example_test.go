package jschema

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Benchmark_buildExample(b *testing.B) {
	s := New("", `{
	"foo": "bar",
	"fizz": [
		1,
		2,
		3
	],
	"buzz": {
		"foo": [
			{"bar": 1},
			{"bar": 2}
		],
		"bar": {
			"fizz": 42,
			"buzz": [1, 2, 3]
		},
		"fizz": 1, // {or: ["string", "integer"]}
		"buzz": 2
	}
}`)
	require.NoError(b, s.Compile())

	node := s.Inner.RootNode()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = buildExample(node, nil)
	}
}
