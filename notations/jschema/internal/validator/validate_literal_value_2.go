package validator

import (
	"fmt"

	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/internal/lexeme"
	"github.com/jsightapi/jsight-schema-go-library/notations/jschema/schema"
)

// ValidateLiteralValue2 used for validating Path variables in the JSight API library
func ValidateLiteralValue2(node schema.Node, rootSchema schema.Schema, jsonValue bytes.Bytes) errors.Error {
	validatorList := NodeValidatorList(node, rootSchema, nil)
	errorsCount := 0

	var err errors.Error

	for _, v := range validatorList {
		err = validateLiteralValue(v.node(), jsonValue)
		if err != nil {
			errorsCount++
		}
	}

	if errorsCount == len(validatorList) {
		if len(validatorList) == 1 {
			return err
		} else {
			return lexeme.NewLexEventError(node.BasisLexEventOfSchemaForNode(), errors.ErrOrRuleSetValidation)
		}
	}

	return nil
}

func validateLiteralValue(node schema.Node, jsonValue bytes.Bytes) (err errors.Error) {
	defer func() {
		if r := recover(); r != nil {
			switch val := r.(type) {
			case errors.DocumentError:
				err = val
			case errors.Err:
				err = lexeme.NewLexEventError(node.BasisLexEventOfSchemaForNode(), val)
			default:
				err = lexeme.NewLexEventError(
					node.BasisLexEventOfSchemaForNode(),
					errors.Format(errors.ErrGeneric, fmt.Sprintf("%s", r)))
			}
		}
	}()

	ValidateLiteralValue(node, jsonValue)

	return nil
}
