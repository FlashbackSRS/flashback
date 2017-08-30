// +build js

// Package file provides information about files and allows JavaScript in a web page to access their content.
//
// See https://developer.mozilla.org/en-US/docs/Web/API/File
package file

import (
	"time"

	"github.com/flimzy/goweb/blob"
	"github.com/gopherjs/gopherjs/js"
)

type File struct {
	blob.Blob
	// Returns the name of the file referenced by the File object. Read only.
	Name string `js:"name"`
}

func InternalizeFile(o *js.Object) *File {
	return &File{Blob: *blob.Internalize(o)}
}

// Returns a time.Time struct representing the last modified time of the file.
func (f *File) LastModified() time.Time {
	ts := f.Get("lastModified").Int64()
	return time.Unix(ts/1000, (ts%1000)*1000000)
}

type FileList struct {
	*js.Object
	// A read-only value indicating the number of files in the list.
	Length int `js:"length"`
}

func InternalizeFileList(o *js.Object) *FileList {
	return &FileList{Object: o}
}

func (l *FileList) Item(idx int) *File {
	return InternalizeFile(l.Call("item", idx))
}
