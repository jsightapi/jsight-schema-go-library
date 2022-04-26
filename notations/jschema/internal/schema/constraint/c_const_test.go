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
		assert.Panics(t, func() {
			fakeConst("foo", "")
		})
	})
}

func TestConst_IsJsonTypeCompatible(t *testing.T) {
	cc := map[json.Type]bool{
		json.TypeObject:  false,
		json.TypeArray:   false,
		json.TypeString:  true,
		json.TypeInteger: true,
		json.TypeFloat:   true,
		json.TypeBoolean: true,
		json.TypeNull:    true,
	}

	for typ, expected := range cc {
		t.Run(typ.String(), func(t *testing.T) {
			assert.Equal(t, expected, Const{}.IsJsonTypeCompatible(typ))
		})
	}
}

func TestConst_Type(t *testing.T) {
	assert.Equal(t, ConstType, Const{}.Type())
}

func TestConst_String(t *testing.T) {
	assert.Equal(t, "const: true", Const{apply: true}.String())
	assert.Equal(t, "const: false", Const{}.String())
}

func TestConst_Bool(t *testing.T) {
	assert.True(t, Const{apply: true}.Bool())
	assert.False(t, Const{}.Bool())
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
		assert.Panics(t, func() {
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
