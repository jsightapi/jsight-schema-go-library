package regex

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/formats/plaintext"
	"github.com/jsightapi/jsight-schema-go-library/fs"
)

func TestNew(t *testing.T) {
	const (
		name    = "foo"
		content = "bar"
	)

	s := New(name, content, WithGeneratorSeed(42))

	assert.Equal(t, fs.NewFile(name, content), s.file)
	assert.Equal(t, int64(42), s.generatorSeed)
	assert.Equal(t, "", s.pattern)
}

func TestFromFile(t *testing.T) {
	file := fs.NewFile("foo", "bar")

	s := FromFile(file, WithGeneratorSeed(42))

	assert.Equal(t, file, s.file)
	assert.Equal(t, int64(42), s.generatorSeed)
	assert.Equal(t, "", s.pattern)
}

func TestSchema_Pattern(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		actual, err := New("", complexRegex).Pattern()
		require.NoError(t, err)
		assert.Equal(t, complexRegexPattern, actual)
	})

	t.Run("negative", func(t *testing.T) {
		_, err := New("", "invalid").Pattern()
		assert.EqualError(t, err, `ERROR (code 1500): Regex should starts with '/' character, but found 'i'
	in line 1 on file 
	> invalid
	--^`)
	})
}

func TestSchema_Len(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		actual, err := New("", complexRegex).Len()
		require.NoError(t, err)
		assert.Equal(t, 430, int(actual))
	})

	t.Run("negative", func(t *testing.T) {
		_, err := New("", "invalid").Len()
		assert.EqualError(t, err, `ERROR (code 1500): Regex should starts with '/' character, but found 'i'
	in line 1 on file 
	> invalid
	--^`)
	})
}

func TestSchema_Example(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]string{
			"/foo/":          "foo",
			"/bar-\\d{0,2}/": "bar-",
			complexRegex:     "d@[228.255.2.a:\\`]",
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				s := New("", given, WithGeneratorSeed(0))
				actual, err := s.Example()
				require.NoError(t, err)

				assert.Equal(t, expected, string(actual))
				assert.True(t, regexp.MustCompile(s.pattern).Match(actual))
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		_, err := New("", "invalid").Len()
		assert.EqualError(t, err, `ERROR (code 1500): Regex should starts with '/' character, but found 'i'
	in line 1 on file 
	> invalid
	--^`)
	})
}

func TestSchema_AddType(t *testing.T) {
	err := (&Schema{}).AddType("foo", nil)
	require.NoError(t, err)
}

func TestSchema_AddRule(t *testing.T) {
	err := (&Schema{}).AddRule("foo", nil)
	require.NoError(t, err)
}

func TestSchema_Check(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		err := New("", complexRegex, WithGeneratorSeed(0)).Check()
		require.NoError(t, err)
	})

	t.Run("negative", func(t *testing.T) {
		err := New("", "invalid").Check()
		assert.EqualError(t, err, `ERROR (code 1500): Regex should starts with '/' character, but found 'i'
	in line 1 on file 
	> invalid
	--^`)
	})
}

func TestSchema_Validate(t *testing.T) {
	s := New("", "/^\\d+$/")
	assert.NoError(t, s.Validate(plaintext.New("", "123")))
	assert.Error(t, s.Validate(plaintext.New("", "abc")))
}

func TestSchema_GetAST(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		actual, err := New("", complexRegex, WithGeneratorSeed(0)).GetAST()
		require.NoError(t, err)
		assert.Equal(t, jschema.ASTNode{
			TokenType:  jschema.TokenTypeString,
			SchemaType: string(jschema.SchemaTypeString),
			Value:      "/" + complexRegexPattern + "/",
		}, actual)
	})

	t.Run("negative", func(t *testing.T) {
		_, err := New("", "invalid").GetAST()
		assert.EqualError(t, err, `ERROR (code 1500): Regex should starts with '/' character, but found 'i'
	in line 1 on file 
	> invalid
	--^`)
	})
}

func TestSchema_UsedUserTypes(t *testing.T) {
	actual, err := (&Schema{}).UsedUserTypes()
	require.NoError(t, err)
	assert.Nil(t, actual)
}

func TestSchema_doCompile(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]string{
			"/foo/":      "foo",
			"/foo/    ":  "foo",
			"/fo\\/o/":   "fo\\/o",
			"/foo\\//":   "foo\\/",
			complexRegex: complexRegexPattern,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				s := New("", given)
				err := s.doCompile()
				require.NoError(t, err)
				assert.Equal(t, expected, s.pattern)
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		cc := map[string]string{
			"foo": `ERROR (code 1500): Regex should starts with '/' character, but found 'f'
	in line 1 on file 
	> foo
	--^`,
			"/foo": `ERROR (code 1501): Regex should ends with '/' character, but found 'o'
	in line 1 on file 
	> /foo
	-----^`,
			"/[-1}/": `ERROR (code 1502): Invalid regex /[-1}/
	in line 1 on file 
	> /[-1}/
	--^`,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				err := New("", given).doCompile()
				assert.EqualError(t, err, expected)
			})
		}
	})
}

const (
	complexRegex        = "/(?:[a-z0-9!#$%&'*+\\/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+\\/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])/"
	complexRegexPattern = "(?:[a-z0-9!#$%&'*+\\/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+\\/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])"
)
