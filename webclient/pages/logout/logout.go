package logout_page

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/flimzy/flashback/clientstate"
	"github.com/flimzy/flashback/webclient/pages"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/console"
)

var jQuery = jquery.NewJQuery

func init() {
	pages.Register("/logout.html", "pagecontainerbeforetransition", BeforeTransition)
}

func BeforeTransition(ctx context.Context, event *jquery.Event, ui *js.Object) pages.Action {
	console.Log("logout BEFORE")

	button := jQuery("#logout")

	button.On("click", func() {
		console.Log("Trying to log out now")
		expireTime, _ := time.Parse(time.RFC3339, "1900-01-01T00:00:00+00:00")
		emptyCookie := &http.Cookie{
			Name:    "AuthSession",
			Expires: expireTime,
		}
		state := ctx.Value("AppState").(*clientstate.State)
		state.Reset()
		js.Global.Get("document").Set("cookie", emptyCookie.String())
		jQuery(":mobile-pagecontainer").Call("pagecontainer", "change", "/")
	})
	return pages.Return()
}
