package loginhandler

import (
	"net/url"

	"github.com/flimzy/jqeventrouter"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/console"
)

var jQuery = jquery.NewJQuery

// BeforeTransition prepares the logout page before display.
func BeforeTransition(providers map[string]string) jqeventrouter.HandlerFunc {
	return func(_ *jquery.Event, _ *js.Object, _ url.Values) bool {
		console.Log("login BEFORE")

		container := jQuery(":mobile-pagecontainer")
		for rel, href := range providers {
			setLoginHandler(container, rel, href)
		}
		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()

		return true
	}
}
