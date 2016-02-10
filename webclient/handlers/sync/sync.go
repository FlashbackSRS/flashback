package sync

import (
	"github.com/flimzy/go-pouchdb"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/console"
	"github.com/flimzy/flashback/util"
)

var jQuery = jquery.NewJQuery

func BeforeTransition(event *jquery.Event, ui *js.Object) bool {
	console.Log("sync BEFORE")

	go func() {
		container := jQuery(":mobile-pagecontainer")
		jQuery("#syncnow", container).On("click", func() {
			console.Log("Attempting to sync something...")
			go DoSync()
		})
		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()
	}()

	return true
}

func DoSync() {
	host := util.CouchHost()
	dbName := "user-" + util.CurrentUser()
	ldb := pouchdb.New(dbName)
	rdb := pouchdb.New(host + "/" + dbName)
	result, err := pouchdb.Replicate(rdb, ldb, pouchdb.Options{})
	console.Log("error = %j", err)
	console.Log("result = %j", result)
}
