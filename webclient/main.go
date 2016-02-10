// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/console"
	"sync"

// 	"golang.org/x/net/context"

// 	"github.com/flimzy/flashback"

	"github.com/flimzy/flashback/util"
// 	"github.com/flimzy/flashback/clientstate"
// 	"github.com/flimzy/flashback/state"
	"github.com/flimzy/go-cordova"
	"github.com/flimzy/jqeventrouter"
	//    "github.com/flimzy/flashback/user"
// 	"github.com/flimzy/flashback/webclient/pages"
// 	_ "github.com/flimzy/flashback/webclient/pages/all"
// 	_ "github.com/flimzy/flashback/webclient/pages/index"
// 	_ "github.com/flimzy/flashback/webclient/pages/login"
// 	_ "github.com/flimzy/flashback/webclient/pages/logout"
// 	_ "github.com/flimzy/flashback/webclient/pages/sync"
// 	_ "github.com/flimzy/flashback/webclient/pages/debug"
	"github.com/flimzy/flashback/webclient/handlers/auth"
	"github.com/flimzy/flashback/webclient/handlers/general"
	"github.com/flimzy/flashback/webclient/handlers/login"
	"github.com/flimzy/flashback/webclient/handlers/logout"
)

// Some spiffy shortcuts
var jQuery = jquery.NewJQuery
var jQMobile *js.Object
var document *js.Object = js.Global.Get("document")

func main() {
	console.Log("in main()")

	var wg sync.WaitGroup

	initjQuery(&wg)
	initCordova(&wg)
// 	state := clientstate.New()
// 	api := flashback.New(jQuery("link[rel=flashback]").Get(0).Get("href").String())
// 	ctx := context.Background()
//	ctx = context.WithValue(ctx, "cordova", cordova)
// 	ctx = context.WithValue(ctx, "AppState", state)
// 	ctx = context.WithValue(ctx, "api", api)
// 	ctx = context.WithValue(ctx, "couchhost", jQuery("link[rel=flashbackdb]").Get(0).Get("href").String())

	// Wait for the above modules to initialize before we initialize jQuery Mobile
	wg.Wait()
	console.Log("Done with main()")
	initjQueryMobile()
}

func initjQuery(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		js.Global.Get("jQuery").Set("cors", true)
	}()
}

func initCordova(wg *sync.WaitGroup) {
	if ! cordova.IsMobile() {
		return
	}
	wg.Add(1)
	document.Call("addEventListener", "deviceready", func() {
		defer wg.Done()
		console.Log("Cordova device ready")
	}, false)
}

func initjQueryMobile() {
	jQuery(document).On("mobileinit", func() {
		console.Log("mobileinit")
		MobileInit()
	})
	// This is what actually loads jQuery Mobile. We have to register our 'mobileinit'
	// event handler above first, though.
	js.Global.Call("postInit")
}

func RouterInit() {
	// beforechange -- Just check auth
	beforeChange := jqeventrouter.NullHandler()
	jqeventrouter.Listen( "pagecontainerbeforechange", general.JQMRouteOnce(general.CleanFacebookURI(auth.CheckAuth(beforeChange))) )

	// beforetransition
	beforeTransition := jqeventrouter.NewEventMux()
	beforeTransition.SetUriFunc( func(_ *jquery.Event, ui *js.Object) string {
		return util.JqmTargetUri(ui)
	})
	beforeTransition.HandleFunc("/login.html", login.BeforeTransition)
	beforeTransition.HandleFunc("/logout.html", logout.BeforeTransition)
	jqeventrouter.Listen( "pagecontainerbeforetransition", beforeTransition )
}

// MobileInit is run after jQuery Mobile's 'mobileinit' event has fired
func MobileInit() {
	jQMobile = js.Global.Get("jQuery").Get("mobile")

	// Disable hash features
	jQMobile.Set("hashListeningEnabled", false)
	jQMobile.Set("pushStateEnabled", false)
	jQMobile.Get("changePage").Get("defaults").Set("changeHash", false)

	DebugEvents()
	RouterInit()

	jQuery(document).On("pagecontainerbeforechange", func(event *jquery.Event, ui *js.Object) {
		console.Log("last beforechange event handler")
	})
	jQuery(document).One("pagecreate", func(event *jquery.Event) {
		console.Log("Enhancing the panel")
		// This should only be executed once, to initialize our "external"
		// panel. This is the kind of thing that should go in document.ready,
		// but I don't have any guarantee that document.ready will run after
		// mobileinit
		jQuery("body>[data-role='panel']").Underlying().Call("panel").Call("enhanceWithin")
	})
	console.Log("Done with MobileInit()")
}

func ConsoleEvent(name string, event *jquery.Event, data *js.Object) {
	page := data.Get("toPage").String()
	if page == "[object Object]" {
		page = data.Get("toPage").Call("jqmData", "url").String()
	}
	console.Log("Event: %s, Current page: %s", name, page)
}

func ConsolePageEvent(name string, event *jquery.Event) {
	console.Log("Event: %s", name)
}

func DebugEvents() {
	events := []string{"pagecontainerbeforehide", "pagecontainerbeforechange", "pagecontainerbeforeload", "pagecontainerbeforeshow",
		"pagecontainerbeforetransition", "pagecontainerchange", "pagecontainerchangefailed", "pagecontainercreate", "pagecontainerhide",
		"pagecontainerload", "pagecontainerloadfailed", "pagecontainerremove", "pagecontainershow", "pagecontainertransition"}
	for _, event := range events {
		copy := event // Necessary for each iterration to have an effectively uinque closure
		jQuery(document).On(event, func(e *jquery.Event, d *js.Object) {
			ConsoleEvent(copy, e, d)
		})
	}
	pageEvents := []string{"beforecreate", "create"}
	for _, event := range pageEvents {
		copy := event
		jQuery(document).On(event, func(e *jquery.Event) {
			ConsolePageEvent(copy, e)
		})
	}
}
