package index_page

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"golang.org/x/net/context"
	"honnef.co/go/js/console"

	"github.com/flimzy/flashback/webclient/pages"
)

func init() {
	pages.Register("/index.html", "pagecontainerbeforetransition", BeforeTransition)
}

func BeforeTransition(ctx context.Context, event *jquery.Event, ui *js.Object) pages.Action {
	console.Log("index BEFORE")

	return pages.Return()
}
