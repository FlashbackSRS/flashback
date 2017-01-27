// +build js

package importhandler

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/url"

	"github.com/flimzy/goweb/file"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/repository"
)

var jQuery = jquery.NewJQuery

// BeforeTransition prepares import page
func BeforeTransition(event *jquery.Event, ui *js.Object, p url.Values) bool {
	go func() {
		container := jQuery(":mobile-pagecontainer")
		jQuery("#importnow", container).On("click", func() {
			fmt.Printf("Foo\n")
			log.Debug("Attempting to import something...\n")
			go func() {
				if err := DoImport(); err != nil {
					log.Printf("Error importing: %s\n", err)
				}
				log.Printf("DoImport() complete\n")
			}()
		})
		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()
	}()

	return true
}

// DoImport does an import of a *.fbb package
func DoImport() error {
	files := file.InternalizeFileList(jQuery("#apkg", ":mobile-pagecontainer").Get(0).Get("files"))
	for i := 0; i < files.Length; i++ {
		if err := importFile(files.Item(i)); err != nil {
			return err
		}
	}
	log.Debugf("Done with import\n")
	return nil
}

func importFile(f *file.File) error {
	u, err := repo.CurrentUser()
	if err != nil {
		return err
	}
	log.Debugf("Gonna import %s now\n", f.Name)
	b, err := f.Bytes()
	if err != nil {
		return err
	}
	z, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer z.Close()
	return repo.Import(u, z)
}
