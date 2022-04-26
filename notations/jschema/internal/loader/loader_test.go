package loader

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/fs"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/scanner"
)

func Test_loadSchema(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		ss := []string{
			`{"key": 1}`,
			"{}",
			"[]",
			`"str"`,
			"123",
			"1.23",
			"true",
			"null",
			"1.2 // {precision: 1}",
		}

		for _, s := range ss {
			t.Run(s, func(t *testing.T) {
				assert.NotPanics(t, func() {
					scan := scanner.NewSchemaScanner(fs.NewFile("", bytes.Bytes(s)), false)
					loadSchema(scan, nil)
				})
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		ss := map[string]string{
			`ERROR (code 301): Invalid character "k" looking for beginning of string
	in line 1 on file 
	> {key: 1}
	---^`: "{key: 1}",
			`ERROR (code 301): Invalid character "," non-space byte after top-level value
	in line 1 on file 
	> 1.2, // {precision: 1}
	-----^`: "1.2, // {precision: 1}",
		}

		for expected, s := range ss {
			t.Run(s, func(t *testing.T) {
				assert.PanicsWithError(t, expected, func() {
					scan := scanner.NewSchemaScanner(fs.NewFile("", bytes.Bytes(s)), false)
					loadSchema(scan, nil)
				})
			})
		}
	})
}
