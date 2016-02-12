package import_handler

import (
// 	"archive/zip"
	"fmt"

// 	"github.com/flimzy/go-pouchdb"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
// 	"github.com/flimzy/flashback/util"
	"honnef.co/go/js/console"
)

var jQuery = jquery.NewJQuery

func BeforeTransition(event *jquery.Event, ui *js.Object) bool {

	go func() {
		container := jQuery(":mobile-pagecontainer")
		jQuery("#importnow", container).On("click", func() {
			fmt.Printf("Attempting to import something...\n")
			go DoImport()
		})
		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()
	}()

	return true
}

func DoImport() {
	files := jQuery("#apkg", ":mobile-pagecontainer").Get(0).Get("files")
	console.Log("%v", files)
	fmt.Printf("files.length = %d\n", files.Length())
//	files := jQuery(":mobile-pagecontainer").Call("getElementById", "apkg")
	for i := 0; i < files.Length(); i++ {
		fmt.Printf("File name: %s\n", files.Index(i).Get("name").String())
	}
// 	fmt.Printf("DoImport()\n")
}

func importFile(file *js.Object) error {
	
}
