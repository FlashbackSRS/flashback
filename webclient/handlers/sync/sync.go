package sync_handler

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/jqeventrouter"
	"github.com/gopherjs/gopherjs/js"

	"github.com/flimzy/flashback/util"
	"github.com/gopherjs/jquery"
)

var jQuery = jquery.NewJQuery
var syncInProgress = false

func SetupSyncButton(h jqeventrouter.Handler) jqeventrouter.Handler {
	return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object, p url.Values) bool {
		fmt.Printf("Setting up the button\n")
		btn := jQuery("[data-id='syncbutton']")
		btn.Off("click")
		btn.On("click", SyncButton)
		if syncInProgress == true {
			disableButton()
		}
		return h.HandleEvent(event, ui, p)
	})
}

func disableButton() {
	syncInProgress = true
	jQuery("[data-id='syncbutton']").AddClass("disabled")
}

func enableButton() {
	syncInProgress = false
	jQuery("[data-id='syncbutton']").RemoveClass("disabled")
}

func SyncButton() {
	fmt.Printf("the button was pressed\n")
	go func() {
		if jQuery("[data-id='syncbutton']").HasClass("disabled") {
			fmt.Printf("button is disabled\n")
			// Sync already in progress
			return
		}
		disableButton()
		DoSync()
		enableButton()
	}()
}

func DoSync() {
	host := util.CouchHost()
	dbName := "user-" + util.CurrentUser()
	ldb := pouchdb.New(dbName)
	rdb := pouchdb.New(host + "/" + dbName)

	var wg sync.WaitGroup
	var docsWritten, docsRead, reviewsWritten int
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		fmt.Printf("Syncing down...\n")
		if docsRead, err = SyncDown(ldb, rdb); err != nil {
			fmt.Printf("Error syncing to server: %s\n", err)
			return
		}
		fmt.Printf("Down sync complete\n")
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		fmt.Printf("Syncing up...\n")
		if docsWritten, err = SyncUp(ldb, rdb); err != nil {
			fmt.Printf("Error syncing from serveR: %s\n", err)
		}
		fmt.Printf("Up sync complete\n")
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		fmt.Printf("Syncing reviews...\n")
		if reviewsWritten, err = SyncReviews(ldb, rdb); err != nil {
			fmt.Printf("Error syncing reviews: %s\n", err)
		}
		fmt.Printf("Review sync complete\n")
	}()
	fmt.Printf("Synced %d docs from server, %d to server, and %d review logs\n", docsRead, docsWritten, reviewsWritten)
	fmt.Printf("Waiting...\n")
	wg.Wait()
	fmt.Printf("Compacting...\n")
	if err := Compact(ldb); err != nil {
		fmt.Printf("Error compacting database: %s\n", err)
		return
	}
	fmt.Printf("Compacting complete\n")
}

func SyncDown(local, remote *pouchdb.PouchDB) (int, error) {
	result, err := pouchdb.Replicate(remote, local, pouchdb.Options{})
	if err != nil {
		return 0, err
	}
	return int(result["docs_written"].(float64)), nil
}

func SyncUp(local, remote *pouchdb.PouchDB) (int, error) {
	result, err := pouchdb.Replicate(local, remote, pouchdb.Options{})
	if err != nil {
		return 0, err
	}
	return int(result["docs_written"].(float64)), nil
}

func SyncReviews(local, remote *pouchdb.PouchDB) (int, error) {
	host := util.CouchHost()
	ldb, err := util.ReviewsSyncDbs()
	if err != nil {
		return 0, err
	}
	if ldb == nil {
		return 0, nil
	}
	before, err := ldb.Info()
	if err != nil {
		return 0, err
	}
	rdb := pouchdb.New(host + "/reviews-" + util.CurrentUser())
	result, err := pouchdb.Replicate(ldb, rdb, pouchdb.Options{})
	if err != nil {
		return 0, err
	}
	revsSynced := int(result["docs_written"].(float64))
	after, err := ldb.Info()
	if err != nil {
		return revsSynced, err
	}
	if before.DocCount != after.DocCount || before.UpdateSeq != after.UpdateSeq {
		fmt.Printf("ReviewsDb content changed during sync. Refusing to delete.\n")
		return revsSynced, nil
	}
	fmt.Printf("Ready to zap %s\n", after.DBName)
	err = util.ZapReviewsDb(ldb)
	return revsSynced, err
}

func Compact(local *pouchdb.PouchDB) error {
	return local.Compact(pouchdb.Options{})
}
