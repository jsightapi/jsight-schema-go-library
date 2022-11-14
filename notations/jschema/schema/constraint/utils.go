package constraint

import (
	"github.com/jsightapi/jsight-schema-go-library/bytes"
	"github.com/jsightapi/jsight-schema-go-library/errors"
)

const colonTrue = ": true"
const colonFalse = ": false"

func parseUint(v bytes.Bytes, c Type) uint {
	u, err := v.ParseUint()
	if err != nil {
		panic(errors.Format(errors.ErrInvalidValueOfConstraint, c.String()))
	}
	return u
}
