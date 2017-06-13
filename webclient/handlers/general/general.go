package general

import (
	"net/url"
	"strings"

	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/util"
)

// CleanFacebookURI cleans up the URL aftr a redirect from a Facebook login
func CleanFacebookURI(h jqeventrouter.Handler) jqeventrouter.Handler {
	return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object, p url.Values) bool {
		uri := util.JqmTargetUri(ui)
		// Having '#_=_' in the URL can mess up our routing
		if strings.HasSuffix(uri, "#_=_") {
			uri = strings.TrimSuffix(uri, "#_=_")
			ui.Set("toPage", uri)
		}
		// TODO: Can we change the hash without refreshing the page?
		// It's also ugly, so remove it from the visible location bar
		// location := js.Global.Get("location")
		// href := location.Get("href").String()
		// if strings.HasSuffix(href, "#_=_") {
		//	location.Set("href", strings.TrimSuffix(href, "#_=_"))
		// }
		return h.HandleEvent(event, ui, p)
	})
}

// JQMRouteOnce ensures that each jQuery Mobile page is only routed once, even when the event is triggered twice (which is common for certain events)
func JQMRouteOnce(h jqeventrouter.Handler) jqeventrouter.Handler {
	return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object, p url.Values) bool {
		if ui.Get("_jqmrouteonce").Bool() {
			log.Debug("pagecontainerbeforechange already ran. Skipping.\n")
			return true
		}
		ui.Set("_jqmrouteonce", true)
		return h.HandleEvent(event, ui, p)
	})
}
