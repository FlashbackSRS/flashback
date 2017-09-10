// +build cordova

package loginhandler

import (
	"net/url"

	"github.com/FlashbackSRS/flashback/model"
	"github.com/flimzy/jqeventrouter"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/console"
)

func setLoginHandler(container jquery.JQuery, rel, href string) {
	li := jQuery("li."+rel, container)
	li.Show()
	jQuery("a", li).On("click", CordovaLogin)
}

// CordovaLogin handles login for the Cordova runtime.
func CordovaLogin() bool {
	console.Log("CordovaLogin()")
	js.Global.Get("facebookConnectPlugin").Call("login", []string{}, func() {
		panic("CordovaLogin needs to be made to work")
		// console.Log("Success logging in")
		// u, err := repo.CurrentUser()
		// if err != nil {
		// 	log.Debugf("No user logged in?? %s\n", err)
		// } else {
		// 	// To make sure the DB is initialized as soon as possible
		// 	u.DB()
		// }
	}, func() {
		console.Log("Failure logging in")
	})
	console.Log("Leaving CordovaLogin()")
	return false
}

// BTCallback does nothing for Cordova.
func BTCallback(_ *model.Repo, _ map[string]string) jqeventrouter.HandlerFunc {
	return func(_ *jquery.Event, _ *js.Object, _ url.Values) bool {
		return true
	}
}
