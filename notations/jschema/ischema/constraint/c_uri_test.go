package constraint

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jsightapi/jsight-schema-go-library/json"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
)

func TestUri_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(t, Uri{}, json.TypeString)
}

func TestUri_Type(t *testing.T) {
	assert.Equal(t, UriConstraintType, NewUri().Type())
}

func TestUri_String(t *testing.T) {
	assert.Equal(t, "uri", NewUri().String())
}

//goland:noinspection ALL
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
				NewUri().Validate(bytes.NewBytes(uri))
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		var cc = map[string]string{
			"":                                "Invalid URI ()",
			"12":                              "Invalid URI (12)",
			"1.2":                             "Invalid URI (1.2)",
			"true":                            "Invalid URI (true)",
			"false":                           "Invalid URI (false)",
			"null":                            "Invalid URI (null)",
			`"ABC"`:                           "Invalid URI (ABC)",
			"example.org":                     "Invalid URI (example.org)",
			" https://example.org":            "Invalid URI ( https://example.org)",
			"/path/to/file.ext":               "Invalid URI (/path/to/file.ext)",
			"path/to/file.ext":                "Invalid URI (path/to/file.ext)",
			"//example.org/path/to/file.ext":  "Invalid URI (//example.org/path/to/file.ext)",
			"://example.org/path/to/file.ext": "Invalid URI (://example.org/path/to/file.ext)",
			"?q=1":                            "Invalid URI (?q=1)",
			"http":                            "Invalid URI (http)",
			"http:":                           "Invalid URI (http:)",
			"http:/":                          "Invalid URI (http:/)",
			"http://":                         "Invalid URI (http://)",
			"http://?q=1":                     "Invalid URI (http://?q=1)",
		}

		for uri, expected := range cc {
			t.Run(uri, func(t *testing.T) {
				assert.PanicsWithError(t, expected, func() {
					NewUri().Validate(bytes.NewBytes(uri))
				})
			})
		}
	})
}

func TestUri_ASTNode(t *testing.T) {
	assert.Equal(t, newEmptyRuleASTNode(), Uri{}.ASTNode())
}
