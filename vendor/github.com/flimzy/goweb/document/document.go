package document

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/flimzy/goweb/element"
)

type Document struct {
	js.Object
}

// The default document
var Document *Document

func init() {
	Document = Internalize(js.Global.Get("document"))
}

func Internalize(o *js.Object) *Document {
	return &Document{Object: *o}
}

func (d *Document) GetElementById(name string) *element.Element {
	return element.Internalize(d.Call("getElementById", string))
}
