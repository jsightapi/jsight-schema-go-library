package test

import (
	"path/filepath"
)

type dir struct {
	relativePath string
	schema       string
	json         []string
	types        []string
	enums        []string
}

func newDir(relativePath string) dir {
	return dir{
		relativePath: relativePath,
		json:         make([]string, 0, 5),
		types:        make([]string, 0, 5),
		enums:        make([]string, 0, 5),
	}
}

func (d dir) isEmpty() bool {
	if d.schema == "" || len(d.json) == 0 {
		return true
	}
	return false
}

func (d *dir) appendFilename(filename string) {
	switch filepath.Ext(filename) {
	case ".jschema":
		d.appendSchema(filename)
	case ".json":
		d.appendJson(filename)
	case ".type":
		d.appendType(filename)
	case ".enum":
		d.appendEnum(filename)
	default:
		panic("Unknown file type: " + filename)
	}
}

func (d *dir) appendSchema(filename string) {
	if d.schema != "" {
		panic("It is possible to have only one schema in the directory: " + d.relativePath)
	}
	d.schema = filename
}

func (d *dir) appendJson(filename string) {
	d.json = append(d.json, filename)
}

func (d *dir) appendType(filename string) {
	d.types = append(d.types, filename)
}

func (d *dir) appendEnum(filename string) {
	d.enums = append(d.enums, filename)
}

func (d dir) String() string {
	return d.relativePath
}
