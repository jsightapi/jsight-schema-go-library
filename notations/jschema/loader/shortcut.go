package loader

import (
	"strings"

	schema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/lexeme"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/ischema"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/ischema/constraint"
)

func addShortcutConstraint(node ischema.Node, rootSchema *ischema.ISchema, lex lexeme.LexEvent) error {
	if lex.Type() != lexeme.TypesShortcutEnd {
		return errors.ErrLoader
	}

	// At this point lexeme value is valid, and we can safely use it.
	// Lexeme value examples:
	// - @foo
	// - @foo | @bar

	// Determines which constraint should be used.
	val := lex.Value().String()

	if strings.ContainsRune(val, '|') {
		addORShortcut(node, rootSchema, val)
	} else {
		addTypeShortcut(node, val)
	}
	return nil
}

func addORShortcut(node ischema.Node, rootSchema *ischema.ISchema, val string) {
	// At this point lexeme value is valid, and we can safely use it.
	// Lexeme value example: "@foo | @bar"

	ss := constraint.NewTypesList(schema.RuleASTNodeSourceGenerated)
	for _, s := range strings.Split(val, "|") {
		typ := ischema.New()
		typ.SetRootNode(node)

		CompileBasic(&typ, true)

		lex := node.BasisLexEventOfSchemaForNode()
		rootSchema.AddUnnamedType(&typ, lex.File(), lex.Begin())

		s = strings.TrimSpace(s)
		ss.AddName(s, s, schema.RuleASTNodeSourceGenerated)
	}

	node.AddConstraint(ss)
	node.AddConstraint(constraint.NewOr(schema.RuleASTNodeSourceGenerated))
}

func addTypeShortcut(node ischema.Node, val string) {
	node.AddConstraint(constraint.NewType(
		bytes.NewBytes(val).TrimSpaces(),
		schema.RuleASTNodeSourceGenerated,
	))
}
