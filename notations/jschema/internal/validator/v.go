package validator

import (
	"j/schema/internal/lexeme"
	"j/schema/notations/jschema/internal/schema"
)

type validator interface {
	parent() validator
	setParent(validator)

	// feed returns array (pointers to validators, or nil if not found), bool
	// (true if validator of node is completed), panic on error.
	feed(jsonLexeme lexeme.LexEvent) ([]validator, bool)

	// node returns this validator node.
	// For debug/log only.
	node() schema.Node

	// log formats this validator for logging.
	// For debug/log only.
	log() string
}
