package sync_handler

import (
	"fmt"
	"net/url"
	"sync"
	"sync/atomic"

	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/jqeventrouter"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/flimzy/flashback/model"
	"github.com/flimzy/flashback/model/user"
	"github.com/flimzy/flashback/util"
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
	u, err := user.CurrentUser()
	if err != nil {
		fmt.Printf("Nobody logged in. Not doing sync\n")
		return
	}
	u.InitDB() // Ensure the Indexes will be built by the time we need them
	dbName := u.DBName()
	ldb := model.NewDB(dbName)
	rdb := model.NewDB(host + "/" + dbName)

	var wg sync.WaitGroup
	var docsWritten, docsRead, reviewsWritten int32
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("Syncing down...\n")
		if r, err := Sync(rdb, ldb); err != nil {
			fmt.Printf("Error syncing to server: %s\n", err)
			return
		} else {
			atomic.AddInt32(&docsRead, r)
		}
		fmt.Printf("Initial down sync complete\n")
		<-u.InitDB() // Make sure the index is built before we try to use it
		if w, r, err := AuxSync(ldb, rdb); err != nil {
			fmt.Printf("Error doing aux sync: %s\n", err)
			return
		} else {
			atomic.AddInt32(&docsRead, r)
			atomic.AddInt32(&docsWritten, w)
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("Syncing up...\n")
		if w, err := Sync(ldb, rdb); err != nil {
			fmt.Printf("Error syncing from serveR: %s\n", err)
		} else {
			atomic.AddInt32(&docsWritten, w)
		}
		fmt.Printf("Up sync complete\n")
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		fmt.Printf("Syncing reviews...\n")
		if rw, err := SyncReviews(ldb, rdb); err != nil {
			fmt.Printf("Error syncing reviews: %s\n", err)
		} else {
			atomic.AddInt32(&reviewsWritten, rw)
		}
		fmt.Printf("Review sync complete\n")
	}()
	wg.Wait()
	fmt.Printf("Synced %d docs from server, %d to server, and %d review logs\n", docsRead, docsWritten, reviewsWritten)
	fmt.Printf("Compacting...\n")
	if err := ldb.Compact(); err != nil {
		fmt.Printf("Error compacting database: %s\n", err)
		return
	}
	fmt.Printf("Compacting complete\n")
}

func Sync(source, target *model.DB) (int32, error) {
	result, err := pouchdb.Replicate(source.PouchDB, target.PouchDB, pouchdb.Options{})
	if err != nil {
		return 0, err
	}
	return int32(result["docs_written"].(float64)), nil
}

func AuxSync(local, remote *model.DB) (int32, int32, error) {
	fmt.Printf("Reading auxilary tables to sync...\n")
	doc := make(map[string][]model.Stub)
	err := local.Find(map[string]interface{}{
		"selector": map[string]string{"type": "stub"},
	}, &doc)
	if err != nil {
		return 0, 0, err
	}
	if len(doc["docs"]) == 0 {
		return 0, 0, nil
	}
	var written, read int32
	host := util.CouchHost()
	var wg sync.WaitGroup
	for _, stub := range doc["docs"] {
		fmt.Printf("stub = %v\n", stub)
		local := model.NewDB(stub.ID)
		remote := model.NewDB(host + "/" + stub.ID)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if w, err := Sync(local, remote); err != nil {
				fmt.Printf("Error syncing '%s' up: %s\n", stub.ID, err)
				return
			} else {
				atomic.AddInt32(&written, w)
			}
		}()
		wg.Add(1)
		go func() {
			defer wg.Done()
			if r, err := Sync(remote, local); err != nil {
				fmt.Printf("Error syncing '%s' down: %s\n", stub.ID, err)
			} else {
				atomic.AddInt32(&read, r)
			}
		}()
	}
	wg.Wait()
	fmt.Printf("Auxilary sync complete\n")
	return written, read, nil
}

func SyncReviews(local, remote *model.DB) (int32, error) {
	u, err := user.CurrentUser()
	if err != nil {
		return 0, err
	}
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
	if before.DocCount == 0 {
		// Nothing at all to sync
		return 0, nil
	}
	rdb := model.NewDB(host + "/" + u.MasterReviewsDBName())
	revsSynced, err := Sync(ldb, rdb)
	if err != nil {
		return 0, err
	}
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
