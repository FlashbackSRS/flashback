// +build js

package synchandler

import (
	"context"
	"net/url"
	"sync/atomic"

	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/kivik"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback/repository"
)

var jQuery = jquery.NewJQuery
var syncInProgress = false

func SetupSyncButton(h jqeventrouter.Handler) jqeventrouter.Handler {
	return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object, p url.Values) bool {
		log.Debugf("Setting up the button\n")
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
	log.Debugf("the button was pressed\n")
	go func() {
		if jQuery("[data-id='syncbutton']").HasClass("disabled") {
			log.Debugf("button is disabled\n")
			// Sync already in progress
			return
		}
		disableButton()
		err := DoSync()
		enableButton()
		if err != nil {
			log.Debugf("Error syncing: %s\n", err)
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
	rdb, err := u.NewRemoteDB(ldb.DBName)
	if err != nil {
		return err
	}

	var docsWritten, docsRead, reviewsWritten int32
	log.Debugf("Syncing down...\n")
	if r, err := Sync(rdb, ldb); err != nil {
		return errors.Errorf("Error syncing to server: %s\n", err)
	} else {
		atomic.AddInt32(&docsRead, r)
	}
	log.Debugf("Down sync complete\n")
	if w, r, err := BundleSync(ldb); err != nil {
		return errors.Errorf("Error syncing bundles: %s\n", err)
	} else {
		log.Debugf("Bundle sync did %d reads, %d writes\n", r, w)
		atomic.AddInt32(&docsRead, r)
		atomic.AddInt32(&docsWritten, w)
	}
	if w, r, err := AuxSync(ldb, rdb); err != nil {
		return errors.Errorf("Error doing aux sync: %s\n", err)
	} else {
		log.Debugf("aux did %d reads, %d writes\n", r, w)
		atomic.AddInt32(&docsRead, r)
		atomic.AddInt32(&docsWritten, w)
	}

	log.Debugf("Syncing up...\n")
	if w, err := Sync(ldb, rdb); err != nil {
		log.Debugf("Error syncing from serveR: %s\n", err)
	} else {
		log.Debugf("up %d docs written\n", w)
		atomic.AddInt32(&docsWritten, w)
	}
	log.Debugf("Up sync complete\n")

	// 	log.Debugf("Syncing reviews...\n")
	// 	if rw, err := SyncReviews(ldb, rdb); err != nil {
	// 		log.Debugf("Error syncing reviews: %s\n", err)
	// 	} else {
	// 		atomic.AddInt32(&reviewsWritten, rw)
	// 	}
	// 	log.Debugf("Review sync complete\n")

	log.Debugf("Synced %d docs from server, %d to server, and %d review logs\n", docsRead, docsWritten, reviewsWritten)
	log.Debugf("Compacting...\n")
	if err := ldb.Compact(context.TODO()); err != nil {
		return errors.Errorf("Error compacting database: %s\n", err)
	}
	log.Debugf("Compacting complete\n")
	return nil
}

// Sync synchronizes local changes with the server
func Sync(source, target *repo.DB) (int32, error) {
	client, _ := kivik.New(context.TODO(), "pouch", "")
	rep, err := client.Replicate(context.TODO(), "", "", map[string]interface{}{
		"source": source.DB,
		"target": target.DB,
	})
	if err != nil {
		return 0, err
	}
	// Just wait until the replication is complete
	// TODO: Visual updates
	for rep.IsActive() {
		rep.Update(context.TODO())
	}
	return int32(rep.DocsWritten()), rep.Err()
}

type bundleResult struct {
	ID string `json:"_id"`
}

// BundleSync syncs auxilary bundles to the remote server.
func BundleSync(udb *repo.DB) (int32, int32, error) {
	log.Debugf("Reading bundles from user database...\n")
	rows, err := udb.Find(context.TODO(), map[string]interface{}{
		"selector": map[string]string{"type": "bundle"},
		"fields":   []string{"_id"},
	})
	if err != nil {
		// FIXME: I'm going to need to find a way to solve this
		// if pouchdb.IsWarning(err) {
		// 	log.Println(err.Error())
		// } else {
		return 0, 0, errors.Wrap(err, "BundleSync failed")
		// }
	}
	var bundles []bundleResult
	for rows.Next() {
		var bundle bundleResult
		if err := rows.ScanDoc(&bundle); err != nil {
			return 0, 0, errors.Wrapf(err, "failed to scan bundle %s", rows.ID())
		}
		bundles = append(bundles, bundle)
	}
	log.Debugf("bundles = %v\n", bundles)
	var written, read int32
	for _, bundle := range bundles {
		log.Debugf("Bundle %s", bundle)
		local, err := udb.User.NewDB(bundle.ID)
		if err != nil {
			return written, read, err
		}
		remote, err := udb.User.NewRemoteDB(bundle.ID)
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
	log.Debugf("Bundles sycned\n")
	return written, read, nil
}

func AuxSync(local, remote *repo.DB) (int32, int32, error) {
	/*
		log.Debugf("Reading auxilary tables to sync...\n")
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
			log.Debugf("stub = %v\n", stub)
			local := model.NewDB(stub.ID)
			remote := model.NewDB(host + "/" + stub.ID)
			if w, err := Sync(local, remote); err != nil {
				log.Debugf("Error syncing '%s' up: %s\n", stub.ID, err)
				return written, read, err
			} else {
				atomic.AddInt32(&written, w)
			}

			if r, err := Sync(remote, local); err != nil {
				log.Debugf("Error syncing '%s' down: %s\n", stub.ID, err)
			} else {
				atomic.AddInt32(&read, r)
			}
		}
		log.Debugf("Auxilary sync complete\n")
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
		log.Debugf("ReviewsDb content changed during sync. Refusing to delete.\n")
		return revsSynced, nil
	}
	log.Debugf("Ready to zap %s\n", after.DBName)
	err = util.ZapReviewsDb(ldb)
	return revsSynced, err
}
*/
