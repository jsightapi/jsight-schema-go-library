package checker

import (
	"j/schema/errors"
	"j/schema/internal/lexeme"
)

type nodeChecker interface {
	check(lexeme.LexEvent) errors.Error
	indentedString(int) string
}
