package auth

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

// 	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/flashback/util"
	"honnef.co/go/js/console"
)

func CheckAuth(h jqeventrouter.Handler) jqeventrouter.Handler {
	return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object) bool {
		console.Log("CheckAuth")
		uri := util.JqmTargetUri(ui)
		console.Log("Auth URI = %s", uri)
		if uri != "/login.html" && util.GetUserFromCookie() == "" {
			console.Log("nobody's logged in")
			// Nobody's logged in
			ui.Set("toPage","login.html")
			event.StopImmediatePropagation()
			console.Log("Attempting to re-trigger the event")
			jquery.NewJQuery(":mobile-pagecontainer").Trigger("pagecontainerbeforechange", ui)
			return true
		}
		console.Log("Auth allowing to proceed")
		return h.HandleEvent(event, ui)
	})
}
