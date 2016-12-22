// +build js

package main

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/flimzy/go-cordova"
	"github.com/flimzy/jqeventrouter"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/fserve"
	"github.com/FlashbackSRS/flashback/util"

	_ "github.com/FlashbackSRS/flashback/models/ankibasic"
	"github.com/FlashbackSRS/flashback/webclient/handlers/auth"
	"github.com/FlashbackSRS/flashback/webclient/handlers/general"
	"github.com/FlashbackSRS/flashback/webclient/handlers/import"
	"github.com/FlashbackSRS/flashback/webclient/handlers/l10n"
	"github.com/FlashbackSRS/flashback/webclient/handlers/login"
	"github.com/FlashbackSRS/flashback/webclient/handlers/logout"
	"github.com/FlashbackSRS/flashback/webclient/handlers/study"
	"github.com/FlashbackSRS/flashback/webclient/handlers/sync"
)

// Some spiffy shortcuts
var jQuery = jquery.NewJQuery
var jQMobile *js.Object
var document *js.Object = js.Global.Get("document")

func main() {
	RouterInit()

	var wg sync.WaitGroup

	initjQuery(&wg)
	initCordova(&wg)
	fserve.Init(&wg)

	// Wait for the above modules to initialize before we initialize jQuery Mobile
	wg.Wait()
	// This is what actually loads jQuery Mobile. We have to register our 'mobileinit'
	// event handler above first, though.
	js.Global.Call("loadjqueryMobile")
}

func initjQuery(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		js.Global.Get("jQuery").Set("cors", true)
		jQuery(js.Global).On("resize", resizeContent)
		jQuery(js.Global).On("orentationchange", resizeContent)
		jQuery(document).On("pagecontainertransition", resizeContent)
	}()
}

func initCordova(wg *sync.WaitGroup) {
	if !cordova.IsMobile() {
		return
	}
	wg.Add(1)
	document.Call("addEventListener", "deviceready", func() {
		defer wg.Done()
	}, false)
}

func resizeContent() {
	screenHt := js.Global.Get("jQuery").Get("mobile").Call("getScreenHeight").Int()
	header := jQuery(".ui-header:visible")
	headerHt := header.OuterHeight()
	if header.HasClass("ui-header-fixed") {
		headerHt = headerHt - 1
	}
	footer := jQuery(".ui-footer:visible")
	footerHt := footer.OuterHeight()
	if footer.HasClass("ui-footer-fixed") {
		footerHt = footerHt - 1
	}
	jQuery(".ui-content").SetHeight(strconv.Itoa(screenHt - headerHt - footerHt))
}

func RouterInit() {
	// mobileinit
	jQuery(document).On("mobileinit", func() {
		MobileInit()
		l10n_handler.MobileInit()
	})

	// beforechange -- Just check auth
	beforeChange := jqeventrouter.NullHandler()
	jqeventrouter.Listen("pagecontainerbeforechange", general.JQMRouteOnce(general.CleanFacebookURI(auth.CheckAuth(beforeChange))))

	// beforetransition
	beforeTransition := jqeventrouter.NewEventMux()
	beforeTransition.SetUriFunc(getJqmUri)
	beforeTransition.HandleFunc("/login.html", loginhandler.BeforeTransition)
	beforeTransition.HandleFunc("/logout.html", logouthandler.BeforeTransition)
	beforeTransition.HandleFunc("/import.html", importhandler.BeforeTransition)
	beforeTransition.HandleFunc("/study.html", studyhandler.BeforeTransition)
	jqeventrouter.Listen("pagecontainerbeforetransition", beforeTransition)

	// beforeshow
	beforeShow := jqeventrouter.NullHandler()
	jqeventrouter.Listen("pagecontainerbeforeshow", l10n_handler.LocalizePage(synchandler.SetupSyncButton(beforeShow)))
}

func getJqmUri(_ *jquery.Event, ui *js.Object) string {
	return util.JqmTargetUri(ui)
}

// MobileInit is run after jQuery Mobile's 'mobileinit' event has fired
func MobileInit() {
	jQMobile = js.Global.Get("jQuery").Get("mobile")

	// Disable hash features
	jQMobile.Set("hashListeningEnabled", false)
	jQMobile.Set("pushStateEnabled", false)
	jQMobile.Get("changePage").Get("defaults").Set("changeHash", false)

	// 	DebugEvents()

	jQuery(document).One("pagecreate", func(event *jquery.Event) {
		// This should only be executed once, to initialize our "external"
		// panel. This is the kind of thing that should go in document.ready,
		// but I don't have any guarantee that document.ready will run after
		// mobileinit
		jQuery("body>[data-role='panel']").Underlying().Call("panel").Call("enhanceWithin")
	})
}

func ConsoleEvent(name string, event *jquery.Event, data *js.Object) {
	page := data.Get("toPage").String()
	if page == "[object Object]" {
		page = data.Get("toPage").Call("jqmData", "url").String()
	}
	fmt.Printf("Event: %s, Current page: %s", name, page)
}

func ConsolePageEvent(name string, event *jquery.Event) {
	fmt.Printf("Event: %s", name)
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
