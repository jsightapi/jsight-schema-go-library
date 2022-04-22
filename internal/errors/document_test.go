package errors

import (
	"j/schema/bytes"
	"j/schema/fs"
	"strings"
	"testing"
)

func TestDetectNewLineSymbol(t *testing.T) {
	f := fs.NewFile("", bytes.Bytes("abc"))
	e := DocumentError{file: f}
	e.preparation()
	if e.nl != '\n' {
		t.Errorf("Incorrect new line symbol %#v", e.nl)
	}

	f = fs.NewFile("", bytes.Bytes("abc\n"))
	e = DocumentError{file: f}
	e.preparation()
	if e.nl != '\n' {
		t.Errorf("Incorrect new line symbol %#v", e.nl)
	}

	f = fs.NewFile("", bytes.Bytes("abc\r\n"))
	e = DocumentError{file: f}
	e.preparation()
	if e.nl != '\n' {
		t.Errorf("Incorrect new line symbol %#v", e.nl)
	}

	f = fs.NewFile("", bytes.Bytes("abc\r"))
	e = DocumentError{file: f}
	e.preparation()
	if e.nl != '\r' {
		t.Errorf("Incorrect new line symbol %#v", e.nl)
	}

	f = fs.NewFile("", bytes.Bytes("abc\n\r"))
	e = DocumentError{file: f}
	e.preparation()
	if e.nl != '\r' {
		t.Errorf("Incorrect new line symbol %#v", e.nl)
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

func TestLineBegin(t *testing.T) {
	for _, d := range data {
		for _, v := range d.valid {
			file := new(fs.File)
			file.SetContent(bytes.Bytes(d.source))

			e := DocumentError{}
			e.SetFile(file)
			e.SetIndex(v.index)
			e.preparation()

			begin := e.lineBeginning()
			if begin != v.begin {
				t.Errorf("Incorrect line beginning [%d] for index [%d] (the expected value is [%d]) on source %#v", begin, v.index, v.begin, d.source)
			}
		}
	}
}

func TestLineEnd(t *testing.T) {
	for _, d := range data {
		for _, v := range d.valid {
			file := new(fs.File)
			file.SetContent(bytes.Bytes(d.source))

			e := DocumentError{}
			e.SetFile(file)
			e.SetIndex(v.index)
			e.preparation()

			end := e.lineEnd()
			if end != v.end {
				t.Errorf("Incorrect line end [%d] for index [%d] (the expected value is [%d]) on source %#v", end, v.index, v.end, d.source)
			}
		}
	}
}

func TestLineNumber(t *testing.T) {
	for _, d := range data {
		for _, v := range d.valid {
			file := new(fs.File)
			file.SetContent(bytes.Bytes(d.source))

			e := DocumentError{}
			e.SetFile(file)
			e.SetIndex(v.index)
			e.preparation()

			n := e.Line()
			if n != v.lineNumber {
				t.Errorf("Incorrect line number [%d] for index [%d] (the expected value is [%d]) on source %#v", n, v.index, v.lineNumber, d.source)
			}
		}
	}
}

func TestSourceString(t *testing.T) {
	for _, d := range data {
		for _, v := range d.valid {
			file := new(fs.File)
			file.SetContent(bytes.Bytes(d.source))

			e := DocumentError{}
			e.SetFile(file)
			e.SetIndex(v.index)
			e.preparation()

			str := e.SourceSubString()
			if str != v.str {
				t.Errorf("Incorrect string %#v for index [%d] (the expected value is \"%s\") on source %#v", str, v.index, v.str, d.source)
			}
		}
	}
}

func TestTooLongSourceSubstringLength(t *testing.T) {
	data := bytes.Bytes(strings.Repeat("123456789 ", 100))

	file := new(fs.File)
	file.SetContent(data)

	e := DocumentError{}
	e.SetFile(file)
	e.SetIndex(0)
	e.preparation()

	if len(e.SourceSubString()) != 200 {
		t.Error("Incorrect length of substring for too long source")
	}
}
