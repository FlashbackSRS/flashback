package auth

import (
	"fmt"
	"log"
	"net/url"

	"github.com/flimzy/jqeventrouter"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/config"
	"github.com/FlashbackSRS/flashback/util"
)

// CheckAuth checks if a uwer has authenticated.
func CheckAuth(c *config.Conf) func(jqeventrouter.Handler) jqeventrouter.Handler {
	return func(h jqeventrouter.Handler) jqeventrouter.Handler {
		return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object, _ url.Values) bool {
			uri := util.JqmTargetUri(ui)
			parsed, err := url.Parse(uri)
			if err != nil {
				log.Printf("Invalid url '%s': %s\n", uri, err)
			}
			if _, ok := parsed.Query()["provider"]; ok {
				if err := repo.Connect(parsed.Query().Get("provider"), parsed.Query().Get("code"), c.GetString("flashback_api")); err != nil {
					fmt.Printf("Failed to authenticate: %s\n", err)
				} else {
					return h.HandleEvent(event, ui, url.Values{})
				}
			}
			if uri != "/app/login.html" && util.CurrentUser() == "" {
				// Nobody's logged in
				ui.Set("toPage", "login.html")
				event.StopImmediatePropagation()
				jquery.NewJQuery(":mobile-pagecontainer").Trigger("pagecontainerbeforechange", ui)
				return true
			}
			return h.HandleEvent(event, ui, url.Values{})
		})
	}
}
