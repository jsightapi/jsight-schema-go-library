package fs

import "j/schema/bytes"

type File struct {
	name     string
	content  bytes.Bytes
	bookmark bytes.Index // for JAPI, in case of INCLUDE file can be read up to a certain point, than left out, then continued
}

func NewFile(name string, content bytes.Bytes) *File {
	return &File{
		name:    name,
		content: content,
	}
}

func (f *File) LastIndex() bytes.Index {
	return bytes.Index(f.Content().Len())
}

func (f *File) Position() bytes.Index {
	return f.bookmark
}

func (f *File) SetPosition(position bytes.Index) {
	f.bookmark = position
}

func (f File) Name() string {
	return f.name
}

func (f *File) SetName(filename string) {
	f.name = filename
}

func (f File) Content() bytes.Bytes {
	return f.content
}

func (f *File) SetContent(content bytes.Bytes) {
	f.content = content
}
