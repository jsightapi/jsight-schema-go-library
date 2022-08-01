package constraint

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

func TestNewConst(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		cc := map[string]bool{
			"true":  true,
			"false": false,
		}

		for given, expected := range cc {
			t.Run(given, func(t *testing.T) {
				c := fakeConst(given, "foo")
				assert.Equal(t, expected, c.apply)
				assert.Equal(t, "foo", string(c.nodeValue))
			})
		}
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, `Invalid value of "const" constraint`, func() {
			fakeConst("foo", "")
		})
	})
}

func TestConst_IsJsonTypeCompatible(t *testing.T) {
	testIsJsonTypeCompatible(
		t,
		Const{},
		json.TypeUndefined,
		json.TypeString,
		json.TypeInteger,
		json.TypeFloat,
		json.TypeBoolean,
		json.TypeNull,
		json.TypeMixed,
	)
}

func TestConst_Type(t *testing.T) {
	assert.Equal(t, ConstType, Const{}.Type())
}

func TestConst_String(t *testing.T) {
	cc := map[string]string{
		"false": "const: false",
		"true":  "const: true",
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			assert.Equal(t, expected, fakeConst(given, "").String())
		})
	}
}

func TestConst_Bool(t *testing.T) {
	cc := map[string]bool{
		"false": false,
		"true":  true,
	}

	for given, expected := range cc {
		t.Run(given, func(t *testing.T) {
			assert.Equal(t, expected, fakeConst(given, "").Bool())
		})
	}
}

func TestConst_Validate(t *testing.T) {
	t.Run("positive", func(t *testing.T) {
		t.Run("apply - true", func(t *testing.T) {
			fakeConst("true", "foo").Validate(bytes.Bytes("foo"))
		})

		t.Run("apply - false", func(t *testing.T) {
			t.Run("valid", func(t *testing.T) {
				fakeConst("false", "foo").Validate(bytes.Bytes("foo"))
			})

			t.Run("invalid", func(t *testing.T) {
				fakeConst("false", "foo").Validate(bytes.Bytes("bar"))
			})
		})
	})

	t.Run("negative", func(t *testing.T) {
		assert.PanicsWithError(t, "Does not match expected value (foo)", func() {
			fakeConst("true", "foo").Validate(bytes.Bytes("bar"))
		})
	})
}

func TestConst_ASTNode(t *testing.T) {
	cc := []bool{true, false}

	for _, c := range cc {
		t.Run(strconv.FormatBool(c), func(t *testing.T) {
			assert.Equal(t, jschema.RuleASTNode{
				JSONType:   jschema.JSONTypeBoolean,
				Value:      strconv.FormatBool(c),
				Properties: &jschema.RuleASTNodes{},
				Source:     jschema.RuleASTNodeSourceManual,
			}, Const{apply: c}.ASTNode())
		})
	}
}

func fakeConst(v, nv string) *Const {
	return NewConst(bytes.Bytes(v), bytes.Bytes(nv))
}
