package loader

import (
	"j/schema/internal/lexeme"
	"j/schema/notations/jschema/internal/scanner"
	"j/schema/notations/jschema/internal/schema"
)

// Contains information about the mode in which the loader is located.
// It affects how the received lexical events will be interpreted depending on whether they are in the comments or not.

type mode int

const (
	readDefault mode = iota
	readInlineComment
	readMultiLineComment
)

// Loads the schema from the scanner into the internal view.
// Does not check for the correctness of the branch because it deals with the scanner.
type loader struct {
	// The schema resulting.
	schema schema.Schema

	// rootSchema a scheme into which types can be added from the "or" rule.
	rootSchema *schema.Schema

	// scanner a tool to search for lexical events in a byte sequence containing
	// a schema.
	scanner *scanner.Scanner

	// mode used for processing inline comment, multi-line comment, or no comment
	// section.
	mode mode

	// nodesPerCurrentLineCount the number of nodes in a line. To check because
	// the rule cannot be added if there is more than one nodes suitable for this
	// in the row.
	nodesPerCurrentLineCount uint

	// lastAddedNode the last node added to the internal Schema.
	lastAddedNode schema.Node

	// The rule is responsible for creating constraints for SCHEMA internal representation
	// nodes from the RULES described in the SCHEMA file.
	rule *ruleLoader

	// The node class is responsible for loading the JSON elements in the nodes
	// of the internal representation of the SCHEMA.
	node *nodeLoader
}

func LoadSchema(scan *scanner.Scanner, rootSchema *schema.Schema, areKeysOptionalByDefault bool) *schema.Schema {
	l := loadSchema(scan, rootSchema)
	CompileBasic(&l.schema, areKeysOptionalByDefault)
	return &(l.schema)
}

func LoadSchemaWithoutCompile(scan *scanner.Scanner, rootSchema *schema.Schema) *schema.Schema {
	return &(loadSchema(scan, rootSchema).schema)
}

func loadSchema(scan *scanner.Scanner, rootSchema *schema.Schema) *loader {
	loader := &loader{
		schema:  schema.New(),
		scanner: scan,
	}

	if rootSchema == nil {
		loader.rootSchema = &loader.schema
	} else {
		loader.rootSchema = rootSchema
	}

	loader.node = newNodeLoader(&loader.schema, &loader.nodesPerCurrentLineCount)

	loader.doLoad()

	return loader
}

// the main function, in which there is a cycle of scanning and loading schemas
func (loader *loader) doLoad() { //nolint:gocyclo // todo try to make this more readable
	for {
		lex, ok := loader.scanner.Next()
		if !ok {
			break
		}

		// useful for debugging comment below 1 line for release
		// fmt.Println("doLoad -> lex", lex)

		switch lex.Type() { //nolint:exhaustive // It's okay here.
		case lexeme.TypesShortcutBegin, lexeme.KeyShortcutBegin:
			continue
		case lexeme.TypesShortcutEnd:
			loader.mode = readDefault
			if err := addShortcutConstraint(loader.lastAddedNode, loader.rootSchema, lex); err != nil {
				panic(err)
			}
			continue
		case lexeme.MultiLineAnnotationBegin:
			loader.mode = readMultiLineComment
			loader.rule = newRuleLoader(loader.lastAddedNode, loader.nodesPerCurrentLineCount, loader.rootSchema)
			continue
		case lexeme.MultiLineAnnotationEnd:
			loader.mode = readDefault
			continue
		case lexeme.InlineAnnotationBegin:
			if loader.mode == readDefault { // not multiLine comment
				loader.mode = readInlineComment
				loader.rule = newRuleLoader(loader.lastAddedNode, loader.nodesPerCurrentLineCount, loader.rootSchema)
				continue
			}
		case lexeme.InlineAnnotationEnd:
			if loader.mode == readInlineComment { // not multiLine comment
				loader.mode = readDefault
				continue
			}
		}

		switch loader.mode {
		case readMultiLineComment, readInlineComment:
			loader.rule.load(lex)
		default:
			if node := loader.node.load(lex); node != nil {
				loader.lastAddedNode = node
			}
		}
	}
}
