// +build js

package importhandler

import (
	"context"
	"net/url"
	"time"

	"github.com/flimzy/goweb/file"
	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/model"
)

var jQuery = jquery.NewJQuery

// BeforeTransition prepares import page
func BeforeTransition(repo *model.Repo) jqeventrouter.HandlerFunc {
	return func(_ *jquery.Event, _ *js.Object, _ url.Values) bool {
		go func() {
			container := jQuery(":mobile-pagecontainer")
			jQuery("#importnow", container).On("click", func() {
				log.Debug("Attempting to import something...\n")
				go func() {
					start := time.Now()
					if err := DoImport(repo); err != nil {
						log.Printf("Error importing: %s\n", err)
					}
					log.Printf("DoImport() complete, %v\n", time.Now().Sub(start))
				}()
			})
			jQuery(".show-until-load", container).Hide()
			jQuery(".hide-until-load", container).Show()
		}()

		return true
	}
}

// DoImport does an import of a *.fbb package
func DoImport(repo *model.Repo) error {
	files := file.InternalizeFileList(jQuery("#apkg", ":mobile-pagecontainer").Get(0).Get("files"))
	for i := 0; i < files.Length; i++ {
		reporter := func(total, progress uint64, percent float64) {
			log.Printf("%d/%d : %.2f\n", progress, total, percent)
		}
		if err := repo.ImportFile(context.TODO(), files.Item(i), reporter); err != nil {
			return err
		}
	}
	log.Debugf("Done with import\n")
	return nil
}
