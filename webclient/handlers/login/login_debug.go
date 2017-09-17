// +build debug

package loginhandler

import (
	"context"
	"fmt"
	"net/url"

	"github.com/FlashbackSRS/flashback/model"
	"github.com/flimzy/jqeventrouter"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

// devLogin handles dev logins in debug builds.
func devLogin(repo *model.Repo) jqeventrouter.HandlerFunc {
	return func(event *jquery.Event, ui *js.Object, params url.Values) bool {
		provider, token := params.Get("provider"), params.Get("token")
		if provider == "" || token == "" {
			displayError("provider and token are both required")
			return true
		}
		go func() {
			if err := repo.Auth(context.TODO(), provider, token); err != nil {
				displayError("dev mode auth failed: " + err.Error())
				return
			}
			fmt.Printf("Dev Mode Auth succeeded!\n")
			ui.Set("toPage", "index.html")
			event.StopImmediatePropagation()
			js.Global.Get("jQuery").Get("mobile").Call("changePage", "index.html")
		}()
		return true
	}
}
