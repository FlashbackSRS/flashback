package auth

import (
	"fmt"
	"log"
	"net/url"

	"github.com/flimzy/jqeventrouter"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/model"
)

// CheckAuth checks if a uwer has authenticated.
func CheckAuth(repo *model.Repo) func(jqeventrouter.Handler) jqeventrouter.Handler {
	return func(h jqeventrouter.Handler) jqeventrouter.Handler {
		return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object, _ url.Values) bool {
			_, err := repo.CurrentUser()
			if err != nil && err != model.ErrNotLoggedIn {
				log.Printf("Unknown error: %s", err)
			}
			if err == model.ErrNotLoggedIn {
				redir := "login.html"
				parsed, _ := url.Parse(js.Global.Get("location").String())
				if p := parsed.Query().Get("provider"); p != "" {
					redir = "callback.html"
				}
				fmt.Printf("Redirecting unauthenticated user to %s\n", redir)
				ui.Set("toPage", redir)
				event.StopImmediatePropagation()
				jquery.NewJQuery(":mobile-pagecontainer").Trigger("pagecontainerbeforechange", ui)
				return true
			}
			return h.HandleEvent(event, ui, url.Values{})
		})
	}
}
