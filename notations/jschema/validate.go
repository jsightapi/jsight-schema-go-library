package jschema

import (
	"fmt"

	jerr "github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/internal/json"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/schema"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/internal/validator"
)

func (s *Schema) ValidateObjectProperty(key, value []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			switch val := r.(type) {
			case jerr.DocumentError:
				err = val
			case jerr.Err:
				err = val
			default:
				err = jerr.Format(jerr.ErrGeneric, fmt.Sprintf("%s", r))
			}
		}
	}()

	root := s.inner.RootNode()
	if root.Type() != json.TypeObject {
		return jerr.ErrObjectExpected
	}

	obj, ok := root.(*schema.ObjectNode)
	if !ok {
		return jerr.ErrObjectExpected
	}

	prop, ok := obj.Child(string(key), false)
	if !ok {
		return jerr.Format(jerr.ErrPropertyNotFound, key)
	}

	validator.ValidateLiteralValue(prop, value) // can panic

	return nil
}
