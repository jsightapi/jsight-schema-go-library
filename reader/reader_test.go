package reader

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/test"
)

func TestRead(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		filename := filepath.Join(test.GetProjectRoot(), "testdata", "examples", "boolean", "boolean.jschema")
		expected := bytes.Bytes(`true // Schema containing a literal example`)

		file := fs.NewFile(filename, expected)

		assert.Equal(t, file, Read(filename))
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, "ERROR: open not_existing_file.jst: no such file or directory", func() {
			Read("not_existing_file.jst")
		})
	})
}
