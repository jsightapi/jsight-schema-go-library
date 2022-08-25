package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/internal/lexeme"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema/constraint"
)

func Test_addShortcutConstraint(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		t.Run("or", func(t *testing.T) {
			const content = "@foo|@bar"

			n := schema.NewNode(lexeme.NewLexEvent(lexeme.ObjectBegin, 0, 0, nil))
			sc := schema.New()
			f := fs.MustNewFile("", content)
			lex := lexeme.NewLexEvent(lexeme.TypesShortcutEnd, 0, bytes.Index(len(content)-1), f)

			err := addShortcutConstraint(n, &sc, lex)
			require.NoError(t, err)

			c := n.Constraint(constraint.TypesListConstraintType)
			require.NotNil(t, c)
			require.IsType(t, &constraint.TypesList{}, c)
			assert.Equal(t, []string{"@foo", "@bar"}, c.(*constraint.TypesList).Names())

			c = n.Constraint(constraint.OrConstraintType)
			require.NotNil(t, c)
			require.IsType(t, &constraint.Or{}, c)
		})

		t.Run("types", func(t *testing.T) {
			const content = "@foo"

			n := schema.NewNode(lexeme.NewLexEvent(lexeme.ObjectBegin, 0, 0, nil))
			sc := schema.New()
			f := fs.MustNewFile("", content)
			lex := lexeme.NewLexEvent(lexeme.TypesShortcutEnd, 0, bytes.Index(len(content)-1), f)

			err := addShortcutConstraint(n, &sc, lex)
			require.NoError(t, err)

			c := n.Constraint(constraint.TypeConstraintType)
			require.NotNil(t, c)
			require.IsType(t, &constraint.TypeConstraint{}, c)

			assert.Equal(t, bytes.Bytes("@foo"), c.(*constraint.TypeConstraint).Bytes())
		})
	})

	t.Run("negative", func(t *testing.T) {
		cc := []lexeme.LexEventType{
			lexeme.LiteralBegin,
			lexeme.LiteralEnd,
			lexeme.ObjectBegin,
			lexeme.ObjectEnd,
			lexeme.ObjectKeyBegin,
			lexeme.ObjectKeyEnd,
			lexeme.ObjectValueBegin,
			lexeme.ObjectValueEnd,
			lexeme.ArrayBegin,
			lexeme.ArrayEnd,
			lexeme.ArrayItemBegin,
			lexeme.ArrayItemEnd,
			lexeme.InlineAnnotationBegin,
			lexeme.InlineAnnotationEnd,
			lexeme.InlineAnnotationTextBegin,
			lexeme.InlineAnnotationTextEnd,
			lexeme.MultiLineAnnotationBegin,
			lexeme.MultiLineAnnotationEnd,
			lexeme.MultiLineAnnotationTextBegin,
			lexeme.MultiLineAnnotationTextEnd,
			lexeme.NewLine,
			lexeme.TypesShortcutBegin,
			lexeme.EndTop,
		}

		for _, c := range cc {
			t.Run(c.String(), func(t *testing.T) {
				err := addShortcutConstraint(nil, nil, lexeme.NewLexEvent(c, 0, 0, nil))
				assert.Equal(t, errors.ErrLoader, err)
			})
		}
	})
}

func Test_addORShortcut(t *testing.T) {
	cc := map[string][]string{
		"@foo":              {"@foo"},
		"\t@foo  \t ":       {"@foo"},
		"@foo | @bar":       {"@foo", "@bar"},
		"\t@foo \t |@bar  ": {"@foo", "@bar"},
	}

	for content, expected := range cc {
		t.Run(content, func(t *testing.T) {
			n := schema.NewNode(lexeme.NewLexEvent(lexeme.ObjectBegin, 0, 0, nil))
			sc := schema.New()

			addORShortcut(n, &sc, content)
			c := n.Constraint(constraint.TypesListConstraintType)
			require.NotNil(t, c)
			require.IsType(t, &constraint.TypesList{}, c)
			assert.Equal(t, expected, c.(*constraint.TypesList).Names())

			c = n.Constraint(constraint.OrConstraintType)
			require.NotNil(t, c)
			require.IsType(t, &constraint.Or{}, c)

			assert.Len(t, sc.TypesList(), len(expected))
		})
	}
}

func Test_addTypeShortcut(t *testing.T) {
	cc := map[string]string{
		"@foo":        "@foo",
		"\t@foo  \t ": "@foo",
	}

	for content, expected := range cc {
		t.Run(content, func(t *testing.T) {
			n := schema.NewNode(lexeme.NewLexEvent(lexeme.ObjectBegin, 0, 0, nil))

			addTypeShortcut(n, content)
			c := n.Constraint(constraint.TypeConstraintType)
			require.NotNil(t, c)
			require.IsType(t, &constraint.TypeConstraint{}, c)

			assert.Equal(t, bytes.Bytes(expected), c.(*constraint.TypeConstraint).Bytes())
		})
	}
}
