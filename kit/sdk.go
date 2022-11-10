package kit

import (
	"fmt"

	"github.com/jsightapi/jsight-schema-go-library/errors"
	"github.com/jsightapi/jsight-schema-go-library/fs"
)

type Error interface {
	Filename() string
	Position() uint
	Line() uint
	Column() uint
	Message() string
	ErrCode() int
	IncorrectUserType() string
}

// ConvertError converts error to Error interface.
func ConvertError(f *fs.File, err any) Error {
	switch e := err.(type) {
	case errors.DocumentError:
		return e
	case errors.Err:
		return errors.NewDocumentError(f, e)
	case error:
		return errors.NewDocumentError(f, errors.Format(errors.ErrGeneric, e.Error()))
	}
	return errors.NewDocumentError(f, errors.Format(errors.ErrGeneric, fmt.Sprintf("%s", err)))
}
