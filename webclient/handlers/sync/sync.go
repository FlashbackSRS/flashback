package sync

import (
	"fmt"
	"github.com/flimzy/go-pouchdb"

	"github.com/flimzy/flashback/util"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

var jQuery = jquery.NewJQuery

func BeforeTransition(event *jquery.Event, ui *js.Object) bool {

	go func() {
		container := jQuery(":mobile-pagecontainer")
		jQuery("#syncnow", container).On("click", func() {
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
	fmt.Printf("Syncing down\n")
	ldb := pouchdb.New(dbName)
	rdb := pouchdb.New(host + "/" + dbName)
	result, err := pouchdb.Replicate(rdb, ldb, pouchdb.Options{})
	if err != nil {
		fmt.Printf("Error syncing from server: %s\n", err)
		return
	}
	docsRead := int(result["docs_written"].(float64))
	fmt.Printf("Syncing up\n")
	result, err = pouchdb.Replicate(ldb, rdb, pouchdb.Options{})
	if err != nil {
		fmt.Printf("Error syncing from server: %s\n", err)
		return
	}
	docsWritten := int(result["docs_written"].(float64))
	fmt.Printf("Synced %d docs from server, and %d to server\n", docsRead, docsWritten)
}
