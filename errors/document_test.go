package errors

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/fs"
)

func TestDocumentError_preparation(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		t.Run("not prepared", func(t *testing.T) {
			e := DocumentError{file: fs.MustNewFile("", "123456")}
			e.preparation()

			assert.EqualValues(t, 6, e.length)
			assert.EqualValues(t, '\n', e.nl)
		})

		t.Run("already prepared", func(t *testing.T) {
			e := DocumentError{prepared: true}
			e.preparation()

			assert.EqualValues(t, 0, e.length)
			assert.EqualValues(t, 0, e.nl)
		})
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithValue(t, "The file is not specified", func() {
			(&DocumentError{}).preparation()
		})
	})
}

func TestDocumentError_detectNewLineSymbol(t *testing.T) {
	cc := map[string]byte{
		"abc":     '\n',
		"abc\n":   '\n',
		"abc\r\n": '\n',
		"abc\r":   '\r',
		"abc\n\r": '\r',
	}

	nameReplacer := strings.NewReplacer("\n", "\\n", "\r", "\\r")

	for given, expected := range cc {
		t.Run(nameReplacer.Replace(given), func(t *testing.T) {
			e := DocumentError{file: fs.MustNewFile("", given)}
			e.detectNewLineSymbol()

			assert.Equal(t, string(expected), string(e.nl))
		})
	}
}

type testValidResult struct {
	index      bytes.Index
	begin      bytes.Index
	end        bytes.Index
	str        string
	lineNumber uint
}
type testData struct {
	source string
	valid  []testValidResult
}

var data = []testData{
	{
		"ABC",
		[]testValidResult{
			{0, 0, 3, "ABC", 1}, // index of the character A
			{1, 0, 3, "ABC", 1}, // index of the character B
			{2, 0, 3, "ABC", 1}, // index of the character C
		},
	},
	{
		"AB\n\nCD\n",
		[]testValidResult{
			{0, 0, 2, "AB", 1}, // index of the character A
			{1, 0, 2, "AB", 1}, // index of the character B
			{2, 0, 2, "AB", 1}, // index of first character "\n"
			{3, 3, 3, "", 2},   // index of second character "\n"
			{4, 4, 6, "CD", 3}, // index of the character C
			{5, 4, 6, "CD", 3}, // index of the character D
			{6, 4, 6, "CD", 3}, // index of third character "\n"
		},
	},
	{
		"AB\r\rCD\r",
		[]testValidResult{
			{0, 0, 2, "AB", 1}, // index of the character A
			{1, 0, 2, "AB", 1}, // index of the character B
			{2, 0, 2, "AB", 1}, // index of first character "\r"
			{3, 3, 3, "", 2},   // index of second character "\r"
			{4, 4, 6, "CD", 3}, // index of the character C
			{5, 4, 6, "CD", 3}, // index of the character D
			{6, 4, 6, "CD", 3}, // index of third character "\r"
		},
	},
	{
		"AB\r\n\r\nCD\r\n",
		[]testValidResult{
			{0, 0, 2, "AB", 1}, // index of the character A
			{1, 0, 2, "AB", 1}, // index of the character B
			{2, 0, 2, "AB", 1}, // index of first character "\r"
			{3, 0, 2, "AB", 1}, // index of first character "\n"
			{4, 4, 4, "", 2},   // index of second character "\r"
			{5, 4, 4, "", 2},   // index of second character "\n"
			{6, 6, 8, "CD", 3}, // index of the character C
			{7, 6, 8, "CD", 3}, // index of the character D
			{8, 6, 8, "CD", 3}, // index of third character "\r"
			{9, 6, 8, "CD", 3}, // index of third character "\n"
		},
	},
	{
		"AB\n\r\n\rCD\n\r",
		[]testValidResult{
			{0, 0, 2, "AB", 1}, // index of the character A
			{1, 0, 2, "AB", 1}, // index of the character B
			{2, 0, 2, "AB", 1}, // index of first character "\r"
			{3, 0, 2, "AB", 1}, // index of first character "\n"
			{4, 4, 4, "", 2},   // index of second character "\r"
			{5, 4, 4, "", 2},   // index of second character "\n"
			{6, 6, 8, "CD", 3}, // index of the character C
			{7, 6, 8, "CD", 3}, // index of the character D
			{8, 6, 8, "CD", 3}, // index of third character "\r"
			{9, 6, 8, "CD", 3}, // index of third character "\n"
		},
	},
	{
		"\n\n\n",
		[]testValidResult{
			{0, 0, 0, "", 1},
			{1, 1, 1, "", 2},
			{2, 2, 2, "", 3},
		},
	},
	{
		"\nA\nB\n",
		[]testValidResult{
			{0, 0, 0, "", 1},
			{1, 1, 2, "A", 2},
			{2, 1, 2, "A", 2},
			{3, 3, 4, "B", 3},
			{4, 3, 4, "B", 3},
		},
	},
}

func TestDocumentError_lineBeginning(t *testing.T) {
	for _, d := range data {
		for _, v := range d.valid {
			t.Run(fmt.Sprintf("%s %d", d.source, v.index), func(t *testing.T) {
				file := fs.MustNewFile("", d.source)

				e := newFakeDocumentError(file, v.index)

				begin := e.lineBeginning()
				assert.Equal(t, v.begin, begin)
			})
		}
	}
}

func TestDocumentError_lineEnd(t *testing.T) {
	for _, d := range data {
		for _, v := range d.valid {
			t.Run(fmt.Sprintf("%s %d", d.source, v.index), func(t *testing.T) {
				file := fs.MustNewFile("", d.source)

				e := newFakeDocumentError(file, v.index)

				end := e.lineEnd()
				assert.Equal(t, v.end, end)
			})
		}
	}
}

func TestNewDocumentError_Line(t *testing.T) {
	for _, d := range data {
		for _, v := range d.valid {
			t.Run(fmt.Sprintf("%s %d", d.source, v.index), func(t *testing.T) {
				file := fs.MustNewFile("", d.source)

				e := newFakeDocumentError(file, v.index)

				n := e.Line()
				assert.Equal(t, v.lineNumber, n)
			})
		}
	}
}

func TestDocumentError_SourceSubString(t *testing.T) {
	for _, d := range data {
		for _, v := range d.valid {
			t.Run(fmt.Sprintf("%s %d", d.source, v.index), func(t *testing.T) {
				file := fs.MustNewFile("", d.source)

				e := newFakeDocumentError(file, v.index)

				str := e.SourceSubString()
				assert.Equal(t, v.str, str)
			})
		}
	}

	t.Run("too long source substring", func(t *testing.T) {
		file := fs.MustNewFile("", strings.Repeat("123456789 ", 100))

		e := newFakeDocumentError(file, 0)

		assert.Len(t, e.SourceSubString(), 200)
	})
}

func newFakeDocumentError(f *fs.File, idx bytes.Index) DocumentError {
	e := DocumentError{}
	e.SetFile(f)
	e.SetIndex(idx)
	e.preparation()
	return e
}
