package enum

import (
	stdErrors "errors"
	"sync"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/internal/lexeme"
	"github.com/jsightapi/jsight-schema-go-library/internal/panics"
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
		e.length, e.computeLengthErr = e.computeLength()
	})
	return e.length, e.computeLengthErr
}

func (e *Enum) computeLength() (length uint, err error) {
	defer func() {
		err = panics.Handle(recover(), err)
	}()

	return newScanner(e.file, scannerComputeLength).Length()
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
	defer func() {
		err = panics.Handle(recover(), err)
	}()

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

type enumChecker struct {
	// stateFunc a function for running a state machine (the current state of the
	// state machine).
	stateFunc func(lexeme.LexEvent)
}

func newEnumChecker() *enumChecker {
	l := &enumChecker{}
	l.stateFunc = l.begin
	return l
}

// Check checks the lexeme sequence to make sure it is an enum. When any error is
// detected, sends DocumentError into a panic.
func (l *enumChecker) Check(lex lexeme.LexEvent) {
	defer lexeme.CatchLexEventError(lex)
	l.stateFunc(lex)
}

// begin of array "["
func (l *enumChecker) begin(lex lexeme.LexEvent) {
	switch lex.Type() {
	case lexeme.NewLine:
		return
	case lexeme.ArrayBegin:
		l.stateFunc = l.arrayItemBeginOrArrayEnd
	default:
		panic(errors.ErrEnumArrayExpected)
	}
}

// arrayItemBeginOrArrayEnd handles beginning of array item begin or the end of array.
// ex: [1 <--
// ex: [" <--
// ex: ] <--
func (l *enumChecker) arrayItemBeginOrArrayEnd(lex lexeme.LexEvent) {
	switch lex.Type() {
	case lexeme.NewLine:
		return
	case lexeme.ArrayItemBegin:
		l.stateFunc = l.literal
	case lexeme.ArrayEnd:
		l.stateFunc = l.afterEndOfEnum
	default:
		panic(errors.ErrLoader)
	}
}

// literal handles the literal value inside array.
func (l *enumChecker) literal(lex lexeme.LexEvent) {
	switch lex.Type() {
	case lexeme.LiteralBegin:
		return
	case lexeme.LiteralEnd:
		l.stateFunc = l.arrayItemEnd
	default:
		panic(errors.ErrIncorrectArrayItemTypeInEnumRule)
	}
}

// arrayItemEnd handles the end of array item.
func (l *enumChecker) arrayItemEnd(lex lexeme.LexEvent) {
	switch lex.Type() {
	case lexeme.NewLine:
		return
	case lexeme.ArrayItemEnd:
		l.stateFunc = l.arrayItemBeginOrArrayEnd
	default:
		panic(errors.ErrLoader)
	}
}

// afterEndOfEnum the method should not be called during normal operation. Ensures
// that the checker will not continue to work after the load is complete.
func (*enumChecker) afterEndOfEnum(lex lexeme.LexEvent) {
	if lex.Type() != lexeme.NewLine {
		panic(errors.ErrUnnecessaryLexemeAfterTheEndOfEnum)
	}
}
