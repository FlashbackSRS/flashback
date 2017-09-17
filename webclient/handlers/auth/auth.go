package auth

import (
	"fmt"
	"net/url"

	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/model"
)

// CheckAuth checks if a user has authenticated.
func CheckAuth(prefix string, repo *model.Repo) func(jqeventrouter.Handler) jqeventrouter.Handler {
	return func(h jqeventrouter.Handler) jqeventrouter.Handler {
		return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object, params url.Values) bool {
			reqURL, _ := url.Parse(ui.Get("toPage").String())
			if reqURL.Path == prefix+"/callback.html" {
				// Allow unauthenticated callback, needed by dev logins
				return true
			}
			_, err := repo.CurrentUser()
			if err != nil && err != model.ErrNotLoggedIn {
				log.Printf("Unknown error: %s", err)
			}
			if err == model.ErrNotLoggedIn {
				redir := "login.html"
				log.Debug("TODO: use params instead of re-parsing URL?")
				parsed, _ := url.Parse(js.Global.Get("location").String())
				fmt.Printf("params = %v\nparsed = %v\n", params, parsed.Query())
				if p := parsed.Query().Get("provider"); p != "" {
					redir = "callback.html"
				}
				log.Printf("Redirecting unauthenticated user to %s\n", redir)
				ui.Set("toPage", redir)
				event.StopImmediatePropagation()
				jquery.NewJQuery(":mobile-pagecontainer").Trigger("pagecontainerbeforechange", ui)
				return true
			}
			return h.HandleEvent(event, ui, url.Values{})
		})
	}
}
