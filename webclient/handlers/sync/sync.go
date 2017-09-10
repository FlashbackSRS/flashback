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
)

var jQuery = jquery.NewJQuery
var syncInProgress = false

func SetupSyncButton(repo *model.Repo) func(jqeventrouter.Handler) jqeventrouter.Handler {
	return func(h jqeventrouter.Handler) jqeventrouter.Handler {
		return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object, p url.Values) bool {
			log.Debugf("Setting up sync button\n")
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
