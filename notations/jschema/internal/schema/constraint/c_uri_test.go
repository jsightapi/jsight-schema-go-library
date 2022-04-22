package constraint

import (
	"github.com/stretchr/testify/assert"
	"j/schema/bytes"
	"testing"
)

func TestUri_Type(t *testing.T) {
	assert.Equal(t, UriConstraintType, NewUri().Type())
}

func TestUri_Validate(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		var tests = []string{
			`scheme://userinfo@host/path?query#fragment`,
			`"https://example.org"`, // json string (with quotes). Will be trimmed before validation.
			`http://example.org`,
			`https://example.org`,
			`ftp://example.org`,
			`ssh://example.org`,
			`http://example.org/path/file.ext?q=1&w=2#zzz`,
			`http://localhost/`,
			`http://localhost:80/`,
			`http://127.0.0.1/`,
			`https://[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:17000`, // IPv6
		}

		for _, uri := range tests {
			t.Run(uri, func(t *testing.T) {
				NewUri().Validate(bytes.Bytes(uri))
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		var tests = []string{
			``,
			`12`,
			`1.2`,
			`true`,
			`false`,
			`null`,
			`"ABC"`,
			`example.org`,
			` https://example.org`, // first space
			`/path/to/file.ext`,    // absolute path
			`path/to/file.ext`,     // relative path
			`//example.org/path/to/file.ext`,
			`://example.org/path/to/file.ext`,
			`?q=1`,
			`http`,
			`http:`,
			`http:/`,
			`http://`,
			`http://?q=1`,
		}

		for _, uri := range tests {
			t.Run(uri, func(t *testing.T) {
				assert.Panics(t, func() {
					NewUri().Validate(bytes.Bytes(uri))
				})
			})
		}
	})
}

func TestUri_ASTNode(t *testing.T) {
	assert.Equal(t, newEmptyRuleASTNode(), Uri{}.ASTNode())
}
