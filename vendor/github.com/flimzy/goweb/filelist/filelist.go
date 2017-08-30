// +build js

// GopherJS Bindings for FileList objects.
//
// See https://developer.mozilla.org/en-US/docs/Web/API/FileList
package filelist

import (
	"github.com/flimzy/goweb/file"
	"github.com/gopherjs/gopherjs/js"
)

type FileList struct {
	*js.Object
	// A read-only value indicating the number of files in the list.
	Length int `js:"length"`
}

func Internalize(o *js.Object) *FileList {
	return &FileList{Object: o}
}

func (l *FileList) Item(idx uint) *file.File {
	return file.Internalize(l.Call("item", idx))
}
