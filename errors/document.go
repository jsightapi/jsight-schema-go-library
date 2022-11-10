package errors

import (
	"fmt"
	"strings"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/fs"
)

// DocumentError contains methods for forming a detailed description of the error
// for a person.
// The resulting message will contain the filename, line number, and where the error
// occurred.
type DocumentError struct {
	// A file containing jSchema or JSON data.
	file              *fs.File
	message           string
	incorrectUserType string
	code              ErrorCode

	// index of the byte in which the error was found.
	index bytes.Index

	line   bytes.Index
	column bytes.Index

	// A length of file content.
	length bytes.Index

	// hasIndex true if the value for Index have been defined.
	hasIndex bool

	// prepared is true when preliminary calculations are made, the results of
	// which are used in some methods.
	prepared bool

	// nl represent new line symbol.
	nl byte
}

var (
	_ Error = DocumentError{}
	_ error = DocumentError{}
)

func NewDocumentError(file *fs.File, err Err) DocumentError {
	return DocumentError{
		code:    err.Code(),
		message: err.Error(),
		file:    file,
	}
}

func (e DocumentError) Code() ErrorCode {
	return e.code
}

func (e DocumentError) ErrCode() int {
	return int(e.code)
}

func (e DocumentError) Filename() string {
	if e.file == nil {
		return ""
	}
	return e.file.Name()
}

func (e DocumentError) Message() string {
	return e.message
}

func (e DocumentError) Position() uint {
	return uint(e.index)
}

func (e DocumentError) Index() bytes.Index {
	return e.index
}

func (e *DocumentError) SetIndex(index bytes.Index) {
	e.index = index
	e.hasIndex = true
}

func (e DocumentError) IncorrectUserType() string {
	return e.incorrectUserType
}

func (e *DocumentError) SetIncorrectUserType(s string) {
	e.incorrectUserType = s
}

func (e *DocumentError) SetFile(file *fs.File) {
	e.file = file
}

func (e *DocumentError) SetMessage(message string) {
	e.message = message
}

// The method performs preparatory calculations, the results of which are used in other methods.
func (e *DocumentError) preparation() {
	if e.prepared {
		return
	}

	if e.file == nil {
		panic("The file is not specified")
	}

	e.length = e.file.Content().LenIndex()
	e.nl = e.file.Content().NewLineSymbol()

	e.prepared = true
}

// lineBeginning
// Before calling this method, you must run the e.preparation().
func (e DocumentError) lineBeginning() bytes.Index {
	content := e.file.Content()
	i := e.index
	if content.LenIndex() <= i {
		return 0
	}
	for {
		c := content.Byte(i)
		if c == e.nl {
			if i != e.index {
				i++ // step forward from new line
				break
			}
		}
		if i == 0 { // It is important because an unsigned value (i := 0; i--; i == [large positive number])
			break
		}
		i--
	}
	return i
}

// lineEnd
// Before calling this method, you must run the e.preparation().
func (e DocumentError) lineEnd() bytes.Index {
	content := e.file.Content()
	i := e.index
	if content.LenIndex() <= i {
		return 0
	}
	for i < e.length {
		c := content.Byte(i)
		if c == e.nl {
			break
		}
		i++
	}
	if i > 0 {
		c := content.Byte(i - 1)
		if (e.nl == '\n' && c == '\r') || (e.nl == '\r' && c == '\n') {
			i--
		}
	}
	return i
}

// Line returns 0 if the line number cannot be determined, or 1+ if it can.
func (e DocumentError) Line() uint {
	if e.line == 0 {
		e.countLineAndColumn()
	}
	return uint(e.line)
}

func (e DocumentError) Column() uint {
	if e.column == 0 {
		e.countLineAndColumn()
	}
	return uint(e.column)
}

func (e *DocumentError) countLineAndColumn() {
	if e.file == nil {
		e.line = 0
		e.column = 0
	} else {
		e.line, e.column = e.file.Content().LineAndColumn(e.index)
	}
}

// SourceSubString returns empty string, if cannot determine the source sub-string.
func (e *DocumentError) SourceSubString() string {
	const maxLength = 200

	if e.file == nil || e.file.Content().Len() == 0 {
		return ""
	}

	e.preparation()

	content := e.file.Content()
	begin := e.lineBeginning()
	end := e.lineEnd()

	if end-begin > maxLength {
		end = begin + maxLength - 3
		return content.Sub(begin, end).TrimSpacesFromLeft().String() + "..."
	}

	return content.Sub(begin, end).TrimSpacesFromLeft().String()
}

func (e *DocumentError) pointerToTheErrorCharacter() string {
	e.preparation()

	content := e.file.Content()
	begin := e.lineBeginning()
	spaces := content.SubLow(begin).CountSpacesFromLeft()

	i := int(e.index) - int(begin) - spaces
	return strings.Repeat("-", i) + "^"
}

func (e DocumentError) Error() string {
	return e.String()
}

func (e *DocumentError) String() string {
	var prefix string
	if e.code == ErrGeneric {
		prefix = "ERROR"
	} else {
		prefix = "ERROR (code " + e.code.Itoa() + ")"
	}
	if e.file != nil {
		filename := e.file.Name()
		if e.hasIndex {
			return fmt.Sprintf(`%s: %s
	in line %d on file %s
	> %s
	--%s`, prefix, e.message, e.Line(), filename, e.SourceSubString(), e.pointerToTheErrorCharacter())
		} else if filename != "" {
			return fmt.Sprintf("%s: %s\n\tin file %s", prefix, e.message, filename)
		}
	}
	return fmt.Sprintf("%s: %s", prefix, e.message)
}
