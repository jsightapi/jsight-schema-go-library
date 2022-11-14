package loader

import (
	"github.com/jsightapi/jsight-schema-go-library/lexeme"
)

type embeddedLoader interface {
	Load(lex lexeme.LexEvent) bool
}
