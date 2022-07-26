package loader

import (
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/internal/lexeme"
)

func newFakeLexEvent(t lexeme.LexEventType) lexeme.LexEvent {
	return lexeme.NewLexEvent(t, 0, 0, nil)
}

func newFakeLexEventWithValue(t lexeme.LexEventType, s string) lexeme.LexEvent {
	f := fs.NewFile("", s)
	return lexeme.NewLexEvent(t, 0, bytes.Index(len(s)-1), f)
}
