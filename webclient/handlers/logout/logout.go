package logouthandler

import (
	"net/url"

	"golang.org/x/net/context"

	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/model"
)

var jQuery = jquery.NewJQuery

// BeforeTransition prepares the logout page before display.
func BeforeTransition(repo *model.Repo) jqeventrouter.HandlerFunc {
	return func(_ *jquery.Event, _ *js.Object, _ url.Values) bool {
		log.Debugf("logout BEFORE\n")

		button := jQuery("#logout")

		button.On("click", func() {
			log.Debugf("Trying to log out now\n")
			if err := repo.Logout(context.TODO()); err != nil {
				log.Printf("Logout failure: %s\n", err)
			}
			jQuery(":mobile-pagecontainer").Call("pagecontainer", "change", "/")
		})
		return true
	}
}
