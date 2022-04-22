package validator

import (
	"j/schema/bytes"
	"j/schema/internal/errors"
	"j/schema/internal/json"
	"j/schema/notations/jschema/internal/schema"
	"j/schema/notations/jschema/internal/schema/constraint"
	"sort"
)

func ValidateLiteralValue(node schema.Node, jsonValue bytes.Bytes) {
	checkNotAnEnum(node, jsonValue)

	// sorting to make it easier to debug the scheme if there are several errors in it
	m := node.ConstraintMap()
	l := m.Len()
	keys := make([]int, 0, l)
	for kv := range m.Iterate() {
		keys = append(keys, int(kv.Key))
	}
	sort.Ints(keys)

	var isNullable bool
	if c, ok := m.Get(constraint.NullableConstraintType); ok {
		isNullable = c.(constraint.BoolKeeper).Bool()
	}

	for _, k := range keys {
		t := constraint.Type(k)
		c := m.GetValue(t)

		if _, ok := c.(*constraint.Enum); ok && isNullable && jsonValue.String() == "null" {
			// Handle cases like `null // {enum: [1, 2], nullable: true}`.
			continue
		}

		if v, ok := c.(constraint.LiteralValidator); ok {
			v.Validate(jsonValue)
		}
	}
}

func checkNotAnEnum(node schema.Node, value bytes.Bytes) {
	if node.Constraint(constraint.EnumConstraintType) != nil {
		return
	}

	jsonType := json.Guess(value).LiteralJsonType() // can panic
	schemaType := node.Type()
	if !(jsonType == schemaType ||
		(jsonType == json.TypeInteger && schemaType == json.TypeFloat) ||
		(jsonType == json.TypeNull && node.Constraint(constraint.NullableConstraintType) != nil)) {
		panic(errors.Format(errors.ErrInvalidValueType, jsonType.String(), schemaType.String()))
	}
}
