package all_pages

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"golang.org/x/net/context"
	"honnef.co/go/js/console"

	"github.com/flimzy/flashback/webclient/pages"
	"github.com/flimzy/flashback/clientstate"
)

func BeforeChange(ctx context.Context, event *jquery.Event, ui *js.Object) pages.Action {
	console.Log("ALL BEFORE")
	state := ctx.Value("AppState").(*clientstate.State)
	target := ctx.Value("target").(string)

	if target != "login.html" {
		if state.User == nil {
			console.Log("No local user defined")
			return pages.Redirect("login.html")
		}
	}
	return pages.Return()
}
