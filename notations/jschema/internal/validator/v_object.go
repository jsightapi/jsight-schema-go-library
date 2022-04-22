package validator

import (
	"bytes"
	"fmt"
	jbytes "j/schema/bytes"
	"j/schema/errors"
	"j/schema/internal/lexeme"
	"j/schema/notations/jschema/internal/schema"
	"j/schema/notations/jschema/internal/schema/constraint"
	"reflect"
	"strings"
)

// Validates json according to jSchema's ObjectNode.

type objectValidator struct {
	// node_ an object or mixed.
	node_   schema.Node
	parent_ validator

	// rootSchema the scheme from which it is possible to receive type by their
	// name.
	rootSchema      schema.Schema
	lastFoundKeyLex lexeme.LexEvent
	requiredKeys    map[string]int
}

func newObjectValidator(node schema.Node, parent validator, rootSchema schema.Schema) *objectValidator {
	switch node.(type) {
	case *schema.ObjectNode, *schema.MixedNode, *schema.MixedValueNode:
		v := objectValidator{
			node_:        node,
			parent_:      parent,
			rootSchema:   rootSchema,
			requiredKeys: make(map[string]int, 5),
		}
		v.initRequiredKeys()
		return &v
	default:
		panic(errors.ErrValidator)
	}
}

func (v *objectValidator) initRequiredKeys() {
	requiredKeysConstraint := v.node_.Constraint(constraint.RequiredKeysConstraintType)
	if requiredKeysConstraint != nil {
		for i, k := range requiredKeysConstraint.(*constraint.RequiredKeys).Keys() {
			v.requiredKeys[k] = i
		}
	}
}

func (v objectValidator) node() schema.Node {
	return v.node_
}

func (v objectValidator) parent() validator {
	return v.parent_
}

func (v *objectValidator) setParent(parent validator) {
	v.parent_ = parent
}

// feed returns array (pointers to validators, or nil if not found) and bool (true
// if validator is done).
func (v *objectValidator) feed(jsonLexeme lexeme.LexEvent) ([]validator, bool) { //nolint:gocyclo // todo do something with that.
	defer lexeme.CatchLexEventError(jsonLexeme)

	switch jsonLexeme.Type() { //nolint:exhaustive // We will throw a panic in over cases.
	case lexeme.ObjectBegin, lexeme.ObjectKeyBegin, lexeme.ObjectValueEnd:
		return nil, false

	case lexeme.ObjectKeyEnd:
		v.lastFoundKeyLex = jsonLexeme
		if _, ok := v.node_.(*schema.ObjectNode); !ok { // mixed node
			panic(lexeme.NewLexEventError(
				v.lastFoundKeyLex,
				errors.Format(errors.ErrSchemaDoesNotSupportKey, v.lastFoundKeyLex.Value().Unquote().String())),
			)
		}
		delete(v.requiredKeys, v.lastFoundKeyLex.Value().Unquote().String())
		return nil, false

	case lexeme.ObjectValueBegin:
		if objectNode, ok := v.node_.(*schema.ObjectNode); ok {
			childNode, ok := objectNode.Child(v.lastFoundKeyLex.Value().Unquote().String())
			if !ok { // child node not found on schema object
				if c := v.node_.Constraint(constraint.RequiredKeysConstraintType); c != nil {
					key, ok := v.validateTypeRules(v.lastFoundKeyLex.Value())
					if ok {
						child, ok := objectNode.Child(key)
						if ok {
							delete(v.requiredKeys, key)
							return NodeValidatorList(child, v.rootSchema, v), false
						}
					}
				}
				if c := v.node_.Constraint(constraint.AdditionalPropertiesConstraintType); c != nil {
					return newAdditionalPropertiesValidator(v.node_, v, c.(*constraint.AdditionalProperties)), false
				}

				panic(lexeme.NewLexEventError(
					v.lastFoundKeyLex,
					errors.Format(errors.ErrSchemaDoesNotSupportKey, v.lastFoundKeyLex.Value().Unquote().String())),
				)
			}
			return NodeValidatorList(childNode, v.rootSchema, v), false
		} else { // mixed node
			panic(errors.ErrImpossible)
		}

	case lexeme.ObjectEnd:
		if len(v.requiredKeys) != 0 {
			panic(errors.Format(errors.ErrRequiredKeyNotFound, v.requiredKeysString()))
		}
		return nil, true
	}

	panic(errors.ErrUnexpectedLexInObjectValidator)
}

func (v objectValidator) requiredKeysString() string {
	keys := make([]string, 0, 5)
	for k := range v.requiredKeys {
		keys = append(keys, k)
	}
	return strings.Join(keys, ", ")
}

func (v objectValidator) log() string {
	return fmt.Sprintf("object [%p]", v.node_)
}

// validate with rules
func (v objectValidator) validateTypeRules(value jbytes.Bytes) (string, bool) {
	for key := range v.requiredKeys {
		if typ, ok := v.rootSchema.TypesList()[key]; ok {
			node := typ.Schema().RootNode()
			if node.Type().String() != "string" {
				panic(errors.Format(errors.ErrInvalidKeyType, v.requiredKeysString()))
			}

			flag := false
			inside := false
			i := 0
			for kv := range node.ConstraintMap().Iterate() {
				inside = true
				if i == 0 {
					flag = true
				}
				flag = flag && checkConstraint(kv.Value, value)
				i++
			}

			if !inside {
				if bytes.Equal(node.Value(), value) {
					flag = true
				}
			}
			if flag {
				// all rules ok for a node
				return key, true
			}
		}
	}
	return "", false
}

func checkConstraint(constr constraint.Constraint, value jbytes.Bytes) (b bool) {
	defer func() {
		if r := recover(); r != nil {
			b = false
		}
	}()

	switch ct := constr.(type) {
	case *constraint.MinLength:
		ct.Validate(value)
		return true
	case *constraint.MaxLength:
		ct.Validate(value)
		return true
	case *constraint.Regex:
		ct.Validate(value)
		return true
	case *constraint.Enum:
		ct.Validate(value)
		return true
	default:
		panic(errors.Format(errors.ErrUnknownRule, reflect.TypeOf(constr)))
	}
}
