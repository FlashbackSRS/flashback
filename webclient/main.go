// +build js

package main

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/flimzy/go-cordova"
	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/config"
	_ "github.com/FlashbackSRS/flashback/fserve" // Load the file server for card assets
	"github.com/FlashbackSRS/flashback/iframes"
	"github.com/FlashbackSRS/flashback/model"
	"github.com/FlashbackSRS/flashback/oauth2"
	"github.com/FlashbackSRS/flashback/util"

	_ "github.com/FlashbackSRS/flashback/controllers/anki" // Anki model controllers
	"github.com/FlashbackSRS/flashback/webclient/handlers/auth"
	"github.com/FlashbackSRS/flashback/webclient/handlers/general"
	"github.com/FlashbackSRS/flashback/webclient/handlers/import"
	"github.com/FlashbackSRS/flashback/webclient/handlers/l10n"
	"github.com/FlashbackSRS/flashback/webclient/handlers/login"
	"github.com/FlashbackSRS/flashback/webclient/handlers/logout"
	"github.com/FlashbackSRS/flashback/webclient/handlers/study"
	synchandler "github.com/FlashbackSRS/flashback/webclient/handlers/sync"
)

// Some spiffy shortcuts
var jQuery = jquery.NewJQuery
var jQMobile *js.Object
var document *js.Object = js.Global.Get("document")

func main() {
	RouterInit()

	var wg sync.WaitGroup

	// Call any async init functions first
	initCordova(&wg)
	// meanwhile, all the synchronous ones
	initjQuery()
	iframes.Init()

	// Wait for the above modules to initialize before we initialize jQuery Mobile
	wg.Wait()

	// This is what actually loads jQuery Mobile. We have to register our 'mobileinit'
	// event handler above first, though, as part of RouterInit
	js.Global.Call("loadjqueryMobile")
}

func initjQuery() {
	js.Global.Get("jQuery").Set("cors", true)
	jQuery(js.Global).On("resize", resizeContent)
	jQuery(js.Global).On("orentationchange", resizeContent)
	jQuery(document).On("pagecontainertransition", resizeContent)
}

func initCordova(wg *sync.WaitGroup) {
	if !cordova.IsMobile() {
		return
	}
	wg.Add(1)
	document.Call("addEventListener", "deviceready", func() {
		// TODO: Don't defer; and perhaps even pass wg.Done directly to the Call method?
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
	confJSON := document.Call("getElementById", "config").Get("innerText").String()
	conf, err := config.NewFromJSON([]byte(confJSON))
	if err != nil {
		panic(err)
	}
	appURL, err := url.Parse(conf.GetString("flashback_app"))
	if err != nil {
		panic(err)
	}
	prefix := strings.TrimSuffix(appURL.Path, "/")

	repo, err := model.New(context.TODO(), conf.GetString("flashback_api"))
	if err != nil {
		panic(err)
	}

	// mobileinit
	jQuery(document).On("mobileinit", func() {
		MobileInit()
		l10n_handler.Init()
	})

	// beforechange -- Just check auth
	beforeChange := jqeventrouter.NullHandler()
	checkAuth := auth.CheckAuth(repo)
	jqeventrouter.Listen("pagecontainerbeforechange", general.JQMRouteOnce(general.CleanFacebookURI(checkAuth(beforeChange))))

	// beforetransition
	beforeTransition := jqeventrouter.NewEventMux()
	beforeTransition.SetUriFunc(getJqmUri)

	providers := map[string]string{
		"facebook": oauth2.FacebookURL(conf.GetString("facebook_client_id"), conf.GetString("flashback_app")),
	}

	beforeTransition.HandleFunc(prefix+"/login.html", loginhandler.BeforeTransition(providers))
	beforeTransition.HandleFunc(prefix+"/callback.html", loginhandler.BTCallback(repo))
	beforeTransition.HandleFunc(prefix+"/logout.html", logouthandler.BeforeTransition(repo))
	beforeTransition.HandleFunc(prefix+"/import.html", importhandler.BeforeTransition(repo))
	beforeTransition.HandleFunc(prefix+"/study.html", studyhandler.BeforeTransition(repo))
	jqeventrouter.Listen("pagecontainerbeforetransition", beforeTransition)

	// beforeshow
	beforeShow := jqeventrouter.NullHandler()
	setupSyncButton := synchandler.SetupSyncButton(repo)
	jqeventrouter.Listen("pagecontainerbeforeshow", l10n_handler.LocalizePage(setupSyncButton(beforeShow)))
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
		// mobileinit.
		jQuery("body>[data-role='panel']").Underlying().Call("panel").Call("enhanceWithin")
	})
}

func ConsoleEvent(name string, event *jquery.Event, data *js.Object) {
	page := data.Get("toPage").String()
	if page == "[object Object]" {
		page = data.Get("toPage").Call("jqmData", "url").String()
	}
	log.Debugf("Event: %s, Current page: %s", name, page)
}

func ConsolePageEvent(name string, event *jquery.Event) {
	log.Debugf("Event: %s", name)
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
