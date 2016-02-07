package login

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/gopherjs/jsbuiltin"
	"honnef.co/go/js/console"

	"github.com/flimzy/flashback"
	"github.com/flimzy/go-cordova"
)

var jQuery = jquery.NewJQuery

func BeforeTransition(event *js.Object) bool {
	console.Log("login BEFORE")
	api := flashback.New()

	go func() {
		providers := api.GetLoginProviders()
		container := jQuery(":mobile-pagecontainer")
		for rel, href := range providers {
			li := jQuery("li."+rel, container)
			li.Show()
			a := jQuery("a", li)
			if cordova.IsMobile() {
				console.Log("Setting on click event")
				a.On("click", CordovaLogin )
			} else {
				a.SetAttr("href", href+"?return="+jsbuiltin.EncodeURIComponent(js.Global.Get("location").Get("href").String()))
			}
		}
		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()
	}()

	return true
}

func CordovaLogin() bool {
	console.Log("CordovaLogin()")
	js.Global.Get("facebookConnectPlugin").Call("login", []string{}, func() {
		console.Log("Success logging in")
	}, func() {
		console.Log("Failure logging in")
	})
	console.Log("Leaving CordovaLogin()")
	return false
}