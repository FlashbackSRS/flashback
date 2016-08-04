package sync_handler

import (
	"errors"
	"fmt"
	"net/url"
	"sync/atomic"

	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/jqeventrouter"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/flimzy/flashback/repository"
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
		err := DoSync()
		enableButton()
		if err != nil {
			fmt.Printf("Error syncing: %s\n", err)
		}
	}()
}

func DoSync() error {
	u, err := repo.CurrentUser()
	if err != nil {
		return errors.New("Nobody logged in. Not doing sync\n")
	}
	ldb, err := u.DB()
	if err != nil {
		return err
	}
	rdb, err := repo.NewRemoteDB(ldb.DBName)
	if err != nil {
		return err
	}

	var docsWritten, docsRead, reviewsWritten int32
	fmt.Printf("Syncing down...\n")
	if r, err := Sync(rdb, ldb); err != nil {
		return fmt.Errorf("Error syncing to server: %s\n", err)
	} else {
		atomic.AddInt32(&docsRead, r)
	}
	fmt.Printf("Down sync complete\n")
	if w, r, err := BundleSync(ldb); err != nil {
		return fmt.Errorf("Error syncing bundles: %s\n", err)
	} else {
		fmt.Printf("Bundle sync did %d reads, %d writes\n", r, w)
		atomic.AddInt32(&docsRead, r)
		atomic.AddInt32(&docsWritten, w)
	}
	if w, r, err := AuxSync(ldb, rdb); err != nil {
		return fmt.Errorf("Error doing aux sync: %s\n", err)
	} else {
		fmt.Printf("aux did %d reads, %d writes\n", r, w)
		atomic.AddInt32(&docsRead, r)
		atomic.AddInt32(&docsWritten, w)
	}

	fmt.Printf("Syncing up...\n")
	if w, err := Sync(ldb, rdb); err != nil {
		fmt.Printf("Error syncing from serveR: %s\n", err)
	} else {
		fmt.Printf("up %d docs written\n", w)
		atomic.AddInt32(&docsWritten, w)
	}
	fmt.Printf("Up sync complete\n")

	// 	fmt.Printf("Syncing reviews...\n")
	// 	if rw, err := SyncReviews(ldb, rdb); err != nil {
	// 		fmt.Printf("Error syncing reviews: %s\n", err)
	// 	} else {
	// 		atomic.AddInt32(&reviewsWritten, rw)
	// 	}
	// 	fmt.Printf("Review sync complete\n")

	fmt.Printf("Synced %d docs from server, %d to server, and %d review logs\n", docsRead, docsWritten, reviewsWritten)
	fmt.Printf("Compacting...\n")
	if err := ldb.Compact(); err != nil {
		return fmt.Errorf("Error compacting database: %s\n", err)
	}
	fmt.Printf("Compacting complete\n")
	return nil
}

func Sync(source, target *repo.DB) (int32, error) {
	result, err := pouchdb.Replicate(source.PouchDB, target.PouchDB, pouchdb.Options{})
	if err != nil {
		return 0, err
	}
	return int32(result["docs_written"].(float64)), nil
}

func BundleSync(udb *repo.DB) (int32, int32, error) {
	fmt.Printf("Reading bundles from user database...\n")
	doc := make(map[string][]map[string]string)
	err := udb.Find(map[string]interface{}{
		"selector": map[string]string{"type": "bundle"},
		"fields":   []string{"_id"},
	}, &doc)
	if err != nil {
		return 0, 0, err
	}
	bundles := make([]string, len(doc["docs"]))
	for i, bundle := range doc["docs"] {
		bundles[i] = bundle["_id"]
	}
	fmt.Printf("bundles = %v\n", bundles)
	var written, read int32
	for _, bundle := range bundles {
		fmt.Printf("Bundle %s", bundle)
		local, err := repo.NewDB(bundle)
		if err != nil {
			return written, read, err
		}
		remote, err := repo.NewRemoteDB(bundle)
		if err != nil {
			return written, read, err
		}
		if w, err := Sync(local, remote); err != nil {
			return written, read, err
		} else {
			atomic.AddInt32(&written, w)
		}

		if r, err := Sync(remote, local); err != nil {
			return written, read, err
		} else {
			atomic.AddInt32(&read, r)
		}
	}
	fmt.Printf("Bundles sycned\n")
	return written, read, nil
}

func AuxSync(local, remote *repo.DB) (int32, int32, error) {
	/*
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
		for _, stub := range doc["docs"] {
			fmt.Printf("stub = %v\n", stub)
			local := model.NewDB(stub.ID)
			remote := model.NewDB(host + "/" + stub.ID)
			if w, err := Sync(local, remote); err != nil {
				fmt.Printf("Error syncing '%s' up: %s\n", stub.ID, err)
				return written, read, err
			} else {
				atomic.AddInt32(&written, w)
			}

			if r, err := Sync(remote, local); err != nil {
				fmt.Printf("Error syncing '%s' down: %s\n", stub.ID, err)
			} else {
				atomic.AddInt32(&read, r)
			}
		}
		fmt.Printf("Auxilary sync complete\n")
		return written, read, nil
	*/
	return 0, 0, nil
}

/*
func SyncReviews(local, remote *repo.DB) (int32, error) {
	u, err := repo.CurrentUser()
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
	rdb, err := repo.NewDB(host + "/" + u.MasterReviewsDBName())
	if err != nil {
		return 0, err
	}
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
*/
