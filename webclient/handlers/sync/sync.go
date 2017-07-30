// +build js

package synchandler

import (
	"context"
	"net/url"

	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/model"
	"github.com/FlashbackSRS/flashback/repository"
)

var jQuery = jquery.NewJQuery
var syncInProgress = false

func SetupSyncButton(repo *model.Repo) func(jqeventrouter.Handler) jqeventrouter.Handler {
	return func(h jqeventrouter.Handler) jqeventrouter.Handler {
		return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object, p url.Values) bool {
			log.Debugf("Setting up the button\n")
			btn := jQuery("[data-id='syncbutton']")
			btn.Off("click")
			btn.On("click", func() {
				SyncButton(repo)
			})
			if syncInProgress == true {
				disableButton()
			}
			return h.HandleEvent(event, ui, p)
		})
	}
}

func disableButton() {
	syncInProgress = true
	jQuery("[data-id='syncbutton']").AddClass("disabled")
}

func enableButton() {
	syncInProgress = false
	jQuery("[data-id='syncbutton']").RemoveClass("disabled")
}

func SyncButton(repo *model.Repo) {
	log.Debugf("the button was pressed\n")
	go func() {
		if jQuery("[data-id='syncbutton']").HasClass("disabled") {
			log.Debugf("button is disabled\n")
			// Sync already in progress
			return
		}
		disableButton()
		err := repo.Sync(context.TODO())
		enableButton()
		if err != nil {
			log.Debugf("Error syncing: %s\n", err)
		}
	}()
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
