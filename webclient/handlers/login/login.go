package login_handler

import (
	"fmt"
	"net/url"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/gopherjs/jsbuiltin"
	"honnef.co/go/js/console"

	"github.com/flimzy/flashback"
	"github.com/flimzy/flashback/repository"
	// 	"github.com/flimzy/flashback-model"
	"github.com/flimzy/go-cordova"
)

var jQuery = jquery.NewJQuery

func BeforeTransition(event *jquery.Event, ui *js.Object, p url.Values) bool {
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
				a.On("click", CordovaLogin)
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
		u, err := repo.CurrentUser()
		if err != nil {
			fmt.Printf("No user logged in?? %s\n", err)
		} else {
			// To make sure the DB is initialized as soon as possible
			u.DB()
		}
	}, func() {
		console.Log("Failure logging in")
	})
	console.Log("Leaving CordovaLogin()")
	return false
}
