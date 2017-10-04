package loginhandler

import (
	"context"
	"net/url"

	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/model"
)

var jQuery = jquery.NewJQuery

// BeforeTransition prepares the logout page before display.
func BeforeTransition(repo *model.Repo, providers map[string]string) jqeventrouter.HandlerFunc {
	return func(_ *jquery.Event, ui *js.Object, _ url.Values) bool {
		log.Debug("login BEFORE")

		cancel := checkLoginStatus(repo)

		container := jQuery(ui.Get("toPage"))
		for rel, href := range providers {
			setLoginHandler(repo, container, rel, href, cancel)
		}
		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()

		return true
	}
}

func checkCtx(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func displayError(msg string) {
	log.Printf("Authentication error: %s\n", msg)
	container := jQuery(":mobile-pagecontainer")
	jQuery("#auth_fail_reason", container).SetText(msg)
	jQuery(".show-until-load", container).Hide()
	jQuery(".hide-until-load", container).Show()
}
