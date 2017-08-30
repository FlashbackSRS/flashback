package element

import (
	"github.com/gopherjs/gopherjs/js"
)

type Element struct {
	js.Object
}

func Internalize(o *js.Object) *Element {
	return &Element{Object: *o}
}
