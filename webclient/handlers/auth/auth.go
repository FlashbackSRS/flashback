package auth

import (
	"net/url"

	"github.com/flimzy/jqeventrouter"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/util"
)

func CheckAuth(h jqeventrouter.Handler) jqeventrouter.Handler {
	return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object, p url.Values) bool {
		uri := util.JqmTargetUri(ui)
		if uri != "/login.html" && util.CurrentUser() == "" {
			// Nobody's logged in
			ui.Set("toPage", "login.html")
			event.StopImmediatePropagation()
			jquery.NewJQuery(":mobile-pagecontainer").Trigger("pagecontainerbeforechange", ui)
			return true
		}
		return h.HandleEvent(event, ui, p)
	})
}
