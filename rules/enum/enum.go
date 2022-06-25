package enum

import (
	stdErrors "errors"
	"sync"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/internal/lexeme"
)

// The Enum rule.
type Enum struct {
	compileErr       error
	computeLengthErr error

	// A file where enum content is placed.
	file *fs.File

	values            []bytes.Bytes
	compileOnce       sync.Once
	computeLengthOnce sync.Once

	length uint
}

var _ jschema.Rule = (*Enum)(nil)

// New creates new Enum rule with specified name and content.
func New(name string, content []byte) *Enum {
	return FromFile(fs.NewFile(name, content))
}

// FromFile creates Enum rule from specified file.
func FromFile(f *fs.File) *Enum {
	return &Enum{file: f}
}

func (e *Enum) Len() (uint, error) {
	e.computeLengthOnce.Do(func() {
		e.length, e.computeLengthErr = newScanner(e.file, scannerComputeLength).Length()
	})
	return e.length, e.computeLengthErr
}

// Check checks that enum is valid.
func (e *Enum) Check() error {
	return e.compile()
}

// Values returns a list of values defined in this enum.
func (e *Enum) Values() ([]bytes.Bytes, error) {
	if err := e.compile(); err != nil {
		return nil, err
	}
	return e.values, nil
}

func (e *Enum) compile() error {
	e.compileOnce.Do(func() {
		e.compileErr = e.doCompile()
	})
	return e.compileErr
}

func (e *Enum) doCompile() (err error) {
	scan := newScanner(e.file)

	for {
		lex, err := scan.Next()
		if stdErrors.Is(err, errEOS) {
			break
		}
		if err != nil {
			return err
		}

		// Collect enum values.
		if lex.Type() == lexeme.LiteralEnd {
			e.values = append(e.values, lex.Value())
		}
	}
	return nil
}
