package reader

import (
	"io/ioutil"
	"j/schema/errors"
	"j/schema/fs"
)

// Read reads the contents of the file, returns a slice of bytes.
func Read(filename string) *fs.File {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		docErr := errors.DocumentError{}
		docErr.SetMessage(err.Error())
		panic(docErr)
	}
	return fs.NewFile(filename, data)
}
