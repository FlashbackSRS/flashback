// +build !debug

package loginhandler

import (
	"net/url"

	"github.com/flimzy/jqeventrouter"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/model"
)

// devLogin does nothing in production builds.
func devLogin(_ *model.Repo) jqeventrouter.HandlerFunc {
	return func(_ *jquery.Event, _ *js.Object, _ url.Values) bool {
		return true
	}
}
