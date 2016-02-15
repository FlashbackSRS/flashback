package import_handler

import (
	"archive/zip"
	"bytes"
	"fmt"

// 	"github.com/flimzy/go-pouchdb"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/flimzy/web/file"
// 	"github.com/flimzy/flashback/util"
// 	"honnef.co/go/js/console"
// 	"github.com/flimzy/web/worker"
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
	for i := 0; i < files.Length(); i++ {
		err := importFile( file.Internalize( files.Index(i) ) )
		if err != nil {
			fmt.Printf("Error importing file: %s\n", err)
		}
	}
}

func importFile(f file.File) error {
	fmt.Printf("Gonna pretend to import %s now\n", f.Name())
	z, err := zip.NewReader( bytes.NewReader( f.Bytes() ), f.Size() )
	if err != nil {
		return err
	}
	for _,file := range z.File {
		fmt.Printf("Archive contains %s\n", file.FileHeader.Name)
		if file.FileHeader.Name == "collection.anki2" {
			// Found the SQLite database
			rc, err := file.Open()
			if err != nil {
				return err
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(rc)
			err = readSQLite(buf.Bytes())
			if err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}
