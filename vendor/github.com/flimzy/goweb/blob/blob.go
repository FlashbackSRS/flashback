// +build js

package blob

import (
	"github.com/gopherjs/gopherjs/js"
)

// Blob wraps a js.Object
type Blob struct {
	*js.Object
	// A boolean value, indicating whether the Blob.close() method has been called on the blob. Closed blobs can not be read. (Read-only)
	IsClosed bool `js:"isClosed"`
	// The size, in bytes, of the data contained in the Blob object. (Read-only)
	Size int64 `js:"size"`
	// A string indicating the MIME type of the data contained in the Blob. If the type is unknown, this string is empty. (Read-only)
	Type string `js:"type"`
}

type Options struct {
	*js.Object
	Type    string `js:"type"`
	Endings string `js:"endings"`
}

// New returns a newly created Blob object whose content consists of the
// concatenation of the array of values given in parameter.
func New(parts []interface{}, opts Options) *Blob {
	blob := js.Global.Get("Blob").New(parts, opts)
	return &Blob{Object: blob}
}

// Internalize internalizes a standard *js.Object to a GlobObj
func Internalize(o *js.Object) *Blob {
	return &Blob{Object: o}
}

// Close closes the blob object, possibly freeing underlying resources.
func (b *Blob) Close() {
	b.Call("close")
}

// Slice returns a new Blob object containing the specified range of bytes of the source BlobObject.
func (b *Blob) Slice(start, end int, contenttype string) *Blob {
	newBlobObject := b.Call("slice", start, end, contenttype)
	return &Blob{
		Object: newBlobObject,
	}
}

// Bytes returns the contents of the Blob as a byte array, or returns an error
func (b *Blob) Bytes() ([]byte, error) {
	done := make(chan struct{})
	fr := js.Global.Get("FileReader").New()
	fr.Set("onloadend", func() { close(done) })
	fr.Call("readAsArrayBuffer", b)
	<-done // Wait for the read to finish
	if err := fr.Get("error"); err != nil {
		return nil, &js.Error{err}
	}
	return js.Global.Get("Uint8Array").New(fr.Get("result")).Interface().([]uint8), nil
}
