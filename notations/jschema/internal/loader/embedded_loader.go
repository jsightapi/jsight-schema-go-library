package loader

import (
	"j/schema/internal/lexeme"
)

type embeddedLoader interface {
	load(lex lexeme.LexEvent) bool
}
