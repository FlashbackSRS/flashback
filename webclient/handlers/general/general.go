package general

import (
	"strings"
	"fmt"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/flimzy/flashback/util"
	"github.com/flimzy/jqeventrouter"
)

func CleanFacebookURI(h jqeventrouter.Handler) jqeventrouter.Handler {
	// This handler cleans up the URL after a redirect from a Facebook login
	return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object) bool {
		uri := util.JqmTargetUri(ui)
		if strings.HasSuffix(uri, "#_=_") {
			ui.Set("toPage", strings.TrimSuffix(uri, "#_=_"))
			js.Global.Get("location").Set("hash", "")
		}
		return h.HandleEvent(event, ui)
	})
}

// JQMRouteOnce ensures that each jQuery Mobile page is only routed once, even when the event is triggered twice (which is common for certain events)
func JQMRouteOnce(h jqeventrouter.Handler) jqeventrouter.Handler {
	return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object) bool {
		if ui.Get("_jqmrouteonce").Bool() {
			fmt.Printf("pagecontainerbeforechange already ran. Skipping.\n")
			return true
		}
		ui.Set("_jqmrouteonce", true)
		return h.HandleEvent(event, ui)
	})
}
