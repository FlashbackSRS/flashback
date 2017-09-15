// +build js

package main

import (
	"context"
	"strconv"
	"sync"

	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/config"
	"github.com/FlashbackSRS/flashback/fserve"
	"github.com/FlashbackSRS/flashback/iframes"
	"github.com/FlashbackSRS/flashback/l10n"
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
	log.Debug("Starting main()\n")
	var wg sync.WaitGroup

	// Call any async init functions first
	initCordova(&wg)
	// meanwhile, all the synchronous ones

	confJSON := document.Call("getElementById", "config").Get("innerText").String()
	conf, err := config.NewFromJSON([]byte(confJSON))
	if err != nil {
		panic(err)
	}

	baseURL := conf.GetString("flashback_app")

	jQuery(document).On("mobileinit", func() {
		MobileInit()
	})

	repo, err := model.New(context.TODO(), conf.GetString("flashback_api"), baseURL)
	if err != nil {
		panic(err)
	}

	fserve.Register(repo)

	initjQuery()
	iframes.Init()

	// Wait for the above modules to initialize before we initialize the router
	// and jQuery Mobile
	wg.Wait()

	// This must happen after cordova init, because it uses cordova to read
	// translation files.
	langSet := l10n_handler.Init(baseURL)

	appPrefix := urlPrefix(baseURL)

	RouterInit(appPrefix, baseURL, conf.GetString("facebook_client_id"), repo, langSet)
	studyhandler.StudyInit(appPrefix)

	// This is what actually loads jQuery Mobile. We have to register our
	//  'mobileinit' event handler above first, though, as part of RouterInit
	js.Global.Call("loadjqueryMobile")
	log.Debug("main() finished\n")
}

func initjQuery() {
	log.Debug("Initializing jQuery\n")
	js.Global.Get("jQuery").Set("cors", true)
	jQuery(js.Global).On("resize", resizeContent)
	jQuery(js.Global).On("orentationchange", resizeContent)
	jQuery(document).On("pagecontainertransition", resizeContent)
	log.Debug("jQuery init complete\n")
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

func RouterInit(prefix, baseURL, facebookID string, repo *model.Repo, langSet *l10n.Set) {
	log.Debug("Initializing router\n")

	// beforechange -- Just check auth
	beforeChange := jqeventrouter.NullHandler()
	checkAuth := auth.CheckAuth(repo)
	jqeventrouter.Listen("pagecontainerbeforechange", general.JQMRouteOnce(general.CleanFacebookURI(checkAuth(beforeChange))))

	// beforetransition
	beforeTransition := jqeventrouter.NewEventMux()
	beforeTransition.SetUriFunc(getJqmUri)

	providers := map[string]string{
		"facebook": oauth2.FacebookURL(facebookID, baseURL),
	}

	beforeTransition.HandleFunc(prefix+"/login.html", loginhandler.BeforeTransition(repo, providers))
	beforeTransition.HandleFunc(prefix+"/callback.html", loginhandler.BTCallback(repo, providers))
	beforeTransition.HandleFunc(prefix+"/logout.html", logouthandler.BeforeTransition(repo))
	beforeTransition.HandleFunc(prefix+"/import.html", importhandler.BeforeTransition(repo))
	beforeTransition.HandleFunc(prefix+"/study.html", studyhandler.BeforeTransition(repo))
	jqeventrouter.Listen("pagecontainerbeforetransition", beforeTransition)

	// beforeshow
	beforeShow := jqeventrouter.NullHandler()
	setupSyncButton := synchandler.SetupSyncButton(repo)
	jqeventrouter.Listen("pagecontainerbeforeshow", l10n_handler.LocalizePage(langSet, setupSyncButton(beforeShow)))
	log.Debug("Router init complete\n")
}

func getJqmUri(_ *jquery.Event, ui *js.Object) string {
	return util.JqmTargetUri(ui)
}

// MobileInit is run after jQuery Mobile's 'mobileinit' event has fired
func MobileInit() {
	log.Debugf("MobileInit")
	defer log.Debugf("MobileInit finished")
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
