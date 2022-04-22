package scanner

import (
	"j/schema/bytes"
	"j/schema/fs"
	"testing"
)

func TestScannerObjectLexeme(t *testing.T) {
	file := new(fs.File)
	file.SetContent(bytes.Bytes(`  {  "key"  :  234  }  `))

	s := NewSchemaScanner(file, false)

	var str string

	objBegin, _ := s.Next()
	str = objBegin.Value().String()
	if str != `{` {
		t.Errorf("Incorrect result: %#v", str)
	}

	keyBegin, _ := s.Next()
	str = keyBegin.Value().String()
	if str != `"` { // opening quote
		t.Errorf("Incorrect result: %#v", str)
	}

	keyEnd, _ := s.Next()
	str = keyEnd.Value().String()
	if str != `"key"` {
		t.Errorf("Incorrect result: %#v", str)
	}

	s.Next() // object value begin

	literalBegin, _ := s.Next()
	str = literalBegin.Value().String()
	if str != `2` { // first character of literal
		t.Errorf("Incorrect result: %#v", str)
	}

	literalEnd, _ := s.Next()
	str = literalEnd.Value().String()
	if str != `234` {
		t.Errorf("Incorrect result: %#v", str)
	}

	s.Next() // object value end

	objectEnd, _ := s.Next()
	str = objectEnd.Value().String()
	if str != `{  "key"  :  234  }` {
		t.Errorf("Incorrect result: %#v", str)
	}
}

func TestScannerArrayLexeme(t *testing.T) {
	file := new(fs.File)
	file.SetContent(bytes.Bytes(`["str",false]`))

	s := NewSchemaScanner(file, false)

	var str string

	arrayBegin, _ := s.Next()
	str = arrayBegin.Value().String()
	if str != `[` {
		t.Errorf("Incorrect result: %#v", str)
	}

	s.Next() // array item begin

	firstLiteralBegin, _ := s.Next()
	str = firstLiteralBegin.Value().String()
	if str != `"` { // opening quote
		t.Errorf("Incorrect result: %#v", str)
	}

	firstLiteralEnd, _ := s.Next()
	str = firstLiteralEnd.Value().String()
	if str != `"str"` {
		t.Errorf("Incorrect result: %#v", str)
	}

	s.Next() // array item end
	s.Next() // array item begin

	secondLiteralBegin, _ := s.Next()
	str = secondLiteralBegin.Value().String()
	if str != `f` { // first character
		t.Errorf("Incorrect result: %#v", str)
	}

	secondLiteralEnd, _ := s.Next()
	str = secondLiteralEnd.Value().String()
	if str != `false` {
		t.Errorf("Incorrect result: %#v", str)
	}

	s.Next() // array item end

	arrayEnd, _ := s.Next()
	str = arrayEnd.Value().String()
	if str != `["str",false]` {
		t.Errorf("Incorrect result: %#v", str)
	}
}

func TestScannerMultiLineCommentLexeme(t *testing.T) {
	file := new(fs.File)
	file.SetContent(bytes.Bytes(`/* {} */`))

	s := NewSchemaScanner(file, false)

	var str string

	multiLineCommentBegin, _ := s.Next()
	str = multiLineCommentBegin.Value().String()
	if str != `/*` {
		t.Errorf("Incorrect result: %#v", str)
	}

	s.Next() // objBegin
	s.Next() // objEnd

	multiLineCommentEnd, _ := s.Next()
	str = multiLineCommentEnd.Value().String()
	if str != `/* {} */` {
		t.Errorf("Incorrect result: %#v", str)
	}
}

func TestScannerInlineCommentLexeme(t *testing.T) {
	file := new(fs.File)
	file.SetContent(bytes.Bytes("123 // { } - some comment\r\n"))

	s := NewSchemaScanner(file, false)

	var str string

	literalBegin, _ := s.Next()
	str = literalBegin.Value().String()
	if str != `1` {
		t.Errorf("Incorrect result: %#v", str)
	}

	literalEnd, _ := s.Next()
	str = literalEnd.Value().String()
	if str != `123` {
		t.Errorf("Incorrect result: %#v", str)
	}

	inlineCommentBegin, _ := s.Next()
	str = inlineCommentBegin.Value().String()
	if str != `//` {
		t.Errorf("Incorrect result: %#v", str)
	}

	objBegin, _ := s.Next()
	str = objBegin.Value().String()
	if str != `{` {
		t.Errorf("Incorrect result: %#v", str)
	}

	objEnd, _ := s.Next()
	str = objEnd.Value().String()
	if str != `{ }` {
		t.Errorf("Incorrect result: %#v", str)
	}

	inlineCommentTextBegin, _ := s.Next()
	str = inlineCommentTextBegin.Value().String()
	if str != `s` {
		t.Errorf("Incorrect result: %#v", str)
	}

	inlineCommentTextEnd, _ := s.Next()
	str = inlineCommentTextEnd.Value().String()
	if str != `some comment` {
		t.Errorf("Incorrect result: %#v", str)
	}

	inlineCommentEnd, _ := s.Next()
	str = inlineCommentEnd.Value().String()
	if str != `// { } - some comment` {
		t.Errorf("Incorrect result: %#v", str)
	}

	newLine1, _ := s.Next()
	str = newLine1.Value().String()
	if str != "\r" {
		t.Errorf("Incorrect result: %#v", str)
	}

	newLine2, _ := s.Next()
	str = newLine2.Value().String()
	if str != "\n" {
		t.Errorf("Incorrect result: %#v", str)
	}
}
