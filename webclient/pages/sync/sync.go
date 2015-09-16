package sync_page

import (
	"golang.org/x/net/context"

	"github.com/flimzy/flashback/webclient/pages"
	"github.com/flimzy/go-pouchdb"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/console"
)

var jQuery = jquery.NewJQuery

func BeforeTransition(ctx context.Context, event *jquery.Event, ui *js.Object) pages.Action {
	console.Log("sync BEFORE")

	go func() {
		container := jQuery(":mobile-pagecontainer")
		jQuery("#syncnow", container).On("click", func() {
			console.Log("Attempting to sync something...")
			go DoSync(ctx)
		})
		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()
	}()

	return pages.Return()
}

func DoSync(ctx context.Context) {
	ldb := pouchdb.New("test-deck")
	rdb := pouchdb.New("https://fbdev.iriscouch.com/test-deck")
	result, err := pouchdb.Replicate(rdb, ldb, pouchdb.Options{})
	console.Log("error = %j", err)
	console.Log("result = %j", result)
}
