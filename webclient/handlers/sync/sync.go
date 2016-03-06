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
	dbName := "user-" + util.CurrentUser()
	ldb := pouchdb.New(dbName)
	rdb := pouchdb.New(host + "/" + dbName)
	docsRead, err := SyncDown(ldb, rdb)
	if err != nil {
		fmt.Printf("Error syncing to server: %s\n", err)
		return
	}
	docsWritten, err := SyncUp(ldb, rdb)
	if err != nil {
		fmt.Printf("Error syncing from serveR: %s\n", err)
	}
	reviewsWritten, err := SyncReviews(ldb, rdb)
	if err != nil {
		fmt.Printf("Error syncing reviews: %s\n", err)
	}
	if err := Compact(ldb); err != nil {
		fmt.Printf("Error compacting database: %s\n", err)
		return
	}
	fmt.Printf("Synced %d docs from server, %d to server, and %d review logs\n", docsRead, docsWritten, reviewsWritten)
}

func SyncDown(local, remote *pouchdb.PouchDB) (int, error) {
	fmt.Printf("Syncing down...\n")
	result, err := pouchdb.Replicate(remote, local, pouchdb.Options{})
	if err != nil {
		return 0, err
	}
	return int(result["docs_written"].(float64)), nil
}

func SyncUp(local, remote *pouchdb.PouchDB) (int, error) {
	fmt.Printf("Syncing up...\n")
	result, err := pouchdb.Replicate(local, remote, pouchdb.Options{})
	if err != nil {
		return 0, err
	}
	return int(result["docs_written"].(float64)), nil
}

func SyncReviews(local, remote *pouchdb.PouchDB) (int, error) {
	fmt.Printf("Syncing reviews...\n")
	host := util.CouchHost()
	ldb, err := util.ReviewsDb()
	if err != nil {
		return 0, err
	}
	rdb := pouchdb.New(host + "/reviews-" + util.CurrentUser())
	result, err := pouchdb.Replicate(ldb, rdb, pouchdb.Options{})
	if err != nil {
		return 0, err
	}
	var all map[string]interface{}
	if err := ldb.AllDocs(&all, pouchdb.Options{}); err != nil {
		return 0, err
	}
	fmt.Printf("all = %v\n", all)
	return int(result["docs_written"].(float64)), nil
}

func Compact(local *pouchdb.PouchDB) error {
	return local.Compact(pouchdb.Options{})
}
