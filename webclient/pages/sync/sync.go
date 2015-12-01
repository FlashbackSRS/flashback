package sync_page

import (
	"golang.org/x/net/context"

	"github.com/flimzy/flashback/clientstate"
	"github.com/flimzy/flashback/webclient/pages"
	"github.com/flimzy/go-pouchdb"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/console"
)

var jQuery = jquery.NewJQuery

func init() {
	pages.Register("/sync.html", "pagecontainerbeforetransition", BeforeTransition)
}

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
	state := ctx.Value("AppState").(*clientstate.State)
	host := ctx.Value("couchhost").(string)
	dbName := "user-" + state.CurrentUser
	ldb := pouchdb.New(dbName)
	rdb := pouchdb.New(host + "/" + dbName)
	result, err := pouchdb.Replicate(rdb, ldb, pouchdb.Options{})
	console.Log("error = %j", err)
	console.Log("result = %j", result)
}
