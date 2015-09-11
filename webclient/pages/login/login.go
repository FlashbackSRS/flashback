package login_page

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/gopherjs/jsbuiltin"
	"golang.org/x/net/context"
	"honnef.co/go/js/console"

	"github.com/flimzy/flashback/webclient/pages"

	"github.com/flimzy/flashback"
)

var jQuery = jquery.NewJQuery
var jQMobile *js.Object
var document *js.Object = js.Global.Get("document")

func BeforeTransition(ctx context.Context, event *jquery.Event, ui *js.Object) pages.Action {
	console.Log("login BEFORE")
	api := ctx.Value("api").(*flashback.FlashbackClient)
	cordova := ctx.Value("cordova").(*js.Object)

	go func() {
		providers := api.GetLoginProviders()
		container := jQuery(":mobile-pagecontainer")
		for rel, href := range providers {
			li := jQuery("li."+rel, container)
			li.Show()
			a := jQuery("a", li)
			if cordova != nil {
				console.Log("Setting on click event")
				a.On("click", func() { CordovaLogin(ctx) })
			} else {
				a.SetAttr("href", href+"?return="+jsbuiltin.EncodeURIComponent(js.Global.Get("location").Get("href").String()))
			}
		}
		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()
		// Remove hidden elements to save a little memory
	}()

	return pages.Return()
}

func CordovaLogin(ctx context.Context) bool {
	console.Log("CordovaLogin()")
	js.Global.Get("facebookConnectPlugin").Call("login", []string{}, func() {
		console.Log("Success logging in")
	}, func() {
		console.Log("Failure logging in")
	})
	console.Log("Leaving CordovaLogin()")
	return false
}
