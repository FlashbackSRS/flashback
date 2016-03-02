package sync_handler

import (
	"fmt"
	"net/url"

	"github.com/flimzy/go-pouchdb"

	"github.com/flimzy/flashback/util"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

var jQuery = jquery.NewJQuery

func BeforeTransition(event *jquery.Event, ui *js.Object, p url.Values) bool {

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
	fmt.Printf("Syncing down\n")
	dbName := "user-" + util.CurrentUser()
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
	fmt.Printf("Compacting\n")
	if err := ldb.Compact(pouchdb.Options{}); err != nil {
		fmt.Printf("Error compacting database: %s\n", err)
		return
	}
	docsWritten := int(result["docs_written"].(float64))
	fmt.Printf("Synced %d docs from server, and %d to server\n", docsRead, docsWritten)
}
