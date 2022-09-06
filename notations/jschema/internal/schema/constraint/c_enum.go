package constraint

import (
	"strings"

	jschema "github.com/jsightapi/jsight-schema-go-library"
	jbytes "github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
)

type Enum struct {
	uniqueIdx map[enumItemValue]struct{}
	ruleName  string
	items     []enumItem
}

type enumItem struct {
	comment string
	enumItemValue
}

type enumItemValue struct {
	value    string
	jsonType json.Type
}

func newEnumItemValue(b jbytes.Bytes) enumItemValue {
	b = b.TrimSpaces()
	t := json.Guess(b).JsonType()
	if t == json.TypeString {
		b = b.Unquote()
	}
	return enumItemValue{value: b.String(), jsonType: t}
}

func (v enumItemValue) String() string {
	if v.jsonType == json.TypeString {
		return `"` + v.value + `"`
	} else {
		return v.value
	}
}

var (
	_ Constraint       = Enum{}
	_ Constraint       = (*Enum)(nil)
	_ LiteralValidator = Enum{}
	_ LiteralValidator = (*Enum)(nil)
)

func NewEnum() *Enum {
	return &Enum{
		uniqueIdx: make(map[enumItemValue]struct{}),
		items:     make([]enumItem, 0, 5),
	}
}

func (Enum) IsJsonTypeCompatible(t json.Type) bool {
	return t.IsLiteralType()
}

func (Enum) Type() Type {
	return EnumConstraintType
}

func (c Enum) String() string {
	var str strings.Builder
	str.WriteString(EnumConstraintType.String())
	str.WriteString(": [")
	for i, v := range c.items {
		str.WriteString(v.enumItemValue.String())
		if len(c.items)-1 != i {
			str.WriteString(", ")
		}
	}
	str.WriteString("]")
	return str.String()
}

func (c *Enum) Append(b jbytes.Bytes) int {
	v := newEnumItemValue(b)
	if _, ok := c.uniqueIdx[v]; ok {
		panic(errors.Format(errors.ErrDuplicationInEnumRule, b.String()))
	}
	idx := len(c.items)
	c.items = append(c.items, enumItem{comment: "", enumItemValue: v})
	c.uniqueIdx[v] = struct{}{}
	return idx
}

func (c *Enum) SetComment(idx int, comment string) {
	c.items[idx].comment = comment
}

func (c *Enum) SetRuleName(s string) {
	c.ruleName = s
}

func (c *Enum) RuleName() string {
	return c.ruleName
}

func (c Enum) Validate(a jbytes.Bytes) {
	v := newEnumItemValue(a)
	for _, b := range c.items {
		if v == b.enumItemValue {
			return
		}
	}
	panic(errors.ErrDoesNotMatchAnyOfTheEnumValues)
}

func (c Enum) ASTNode() jschema.RuleASTNode {
	const source = jschema.RuleASTNodeSourceManual

	if c.ruleName != "" {
		return newRuleASTNode(jschema.TokenTypeShortcut, c.ruleName, source)
	}

	n := newRuleASTNode(jschema.TokenTypeArray, "", source)
	n.Items = make([]jschema.RuleASTNode, 0, len(c.items))

	for _, b := range c.items {
		an := newRuleASTNode(
			b.jsonType.ToTokenType(),
			b.value,
			source,
		)
		an.Comment = b.comment

		n.Items = append(n.Items, an)
	}

	return n
}
