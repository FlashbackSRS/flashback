package pages

import (
	"net/url"
	"strings"

	"github.com/flimzy/flashback/clientstate"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"golang.org/x/net/context"
	"honnef.co/go/js/console"
)

var jQuery = jquery.NewJQuery
var document *js.Object = js.Global.Get("document")
var jQMobile *js.Object

type Action struct {
	Action      string
	RedirectUrl string
}

type HandlerFunc func(context.Context, *jquery.Event, *js.Object) Action

func Return() Action {
	return Action{"return", ""}
}

func Next() Action {
	return Action{"next", ""}
}

func Redirect(url string) Action {
	return Action{"redirect", url}
}

type Router struct {
	handlers map[string]map[string]HandlerFunc
}

// Router singleton
var routerSingleton *Router

func init() {
	routerSingleton = &Router{
			make(map[string]map[string]HandlerFunc),
		}
}

func Register(target, event string, pageHandler HandlerFunc) {
	r := routerSingleton
console.Log("Register: %s", target)
	if _, ok := r.handlers[event]; !ok {
		r.handlers[event] = make(map[string]HandlerFunc)
	}
	if _, ok := r.handlers[event][target]; ok {
		panic(event + " handler already registered for " + target)
	}
	r.handlers[event][target] = pageHandler
}

func Init(ctx context.Context) {
	r := routerSingleton
	jQMobile = js.Global.Get("jQuery").Get("mobile")

	for event, _ := range r.handlers {
		name := event
		console.Log("Found handlers for event '%s'", event)
		if event == "pagecontainerbeforechange" {
			// A one-off ugly hack because our router expects toPage to be a string on the first call
			// but normally on the first page load, it is already a parsed DOM object
			jQuery(document).One("pagecontainerbeforechange", func(event *jquery.Event, ui *js.Object) {
				ui.Set("toPage", js.Global.Get("location").Get("href").String())
			})
			jQuery(document).On(name, func(e *jquery.Event, ui *js.Object) {
				r.BeforeChangeRouter(ctx, e, ui)
			})
		} else {
			jQuery(document).On(name, func(e *jquery.Event, ui *js.Object) {
				r.EventRouter(ctx, name, e, ui)
			})
		}
	}
}

func pageName(ui *js.Object) string {
	rawUrl := ui.Get("toPage").String()
	if rawUrl == "[object Object]" {
		rawUrl = ui.Get("toPage").Call("jqmData", "url").String()
	}
	pageUrl, _ := url.Parse(rawUrl)
	path := strings.TrimPrefix(pageUrl.Path, "/android_asset/www")
	if len(pageUrl.Fragment) > 0 {
		return path + "#" + pageUrl.Fragment
	}
	return path
}

func (r *Router) EventRouter(ctx context.Context, eventName string, event *jquery.Event, ui *js.Object) {
	var page string
	if eventName == "pagecreate" || eventName == "pagebeforecreate" {
		page = jQMobile.Call("activePage").String()
	} else {
		page = pageName(ui)
	}
	console.Log("Event %s was routed to me for page %s!!", eventName, page)
	targets := [3]string{"BEFORE", page, "AFTER"}
	for _, target := range targets {
		if handler, ok := r.handlers[eventName][target]; ok {
			handler(ctx, event, ui)
			return
		}
	}
}

func (r *Router) BeforeChangeRouter(ctx context.Context, event *jquery.Event, ui *js.Object) {
	if ui.Get("_pbc").Bool() {
		console.Log("pagecontainerbeforechange already ran. Skipping")
		return
	}
	ui.Set("_pbc", true)
	page := pageName(ui)
	if strings.HasSuffix(page, "#_=_") {
		// This happens after a redirect from FaceBook login
		page = strings.TrimSuffix(page, "#_=_")
		ui.Set("toPage", page)
		js.Global.Get("location").Set("hash", "")
	}

	console.Log("Routing %s", page)

	state := ctx.Value("AppState").(*clientstate.State)
	// See if we're in a redirection, and if so fetch the last context
	if len(state.Stack) > 0 {
		if len(state.Stack) > 5 {
			console.Log("Redirect loop!!!!!")
			ui.Set("toPage", "fatal.html")
			return
		}
		newCtx := state.Stack[len(state.Stack)-1]
		lastPage := newCtx.Value("target").(string)
		console.Log("page = %s, newPage = %s\n", page, lastPage)
		if lastPage != page {
			// This means we aren't in a redirect iteration, so we need to
			// restore context
			ctx = newCtx
		}
	}

	console.Log("Storing %s to ctx.target", page)
	ctx = context.WithValue(ctx, "target", page)
	targets := [3]string{"BEFORE", page, "AFTER"}
	for _, target := range targets {
		if handler, ok := r.handlers["pagecontainerbeforechange"][target]; ok {
			console.Log("Found handler for %s", target)
			result := handler(ctx, event, ui)
			if result.Action == "redirect" {
				ui.Set("toPage", result.RedirectUrl)
				// Store the current context state for the next request
				state.Stack = append(state.Stack, ctx)
				console.Log("Cancelling event propogation")
				console.Log("Event target: %s", event.Target)
				event.StopImmediatePropagation()
				console.Log("Attempting to re-trigger the event")
				jQuery("body.ui-mobile-viewport").Trigger("pagecontainerbeforechange", ui)
				return
			}
			if result.Action == "return" {
				// Do something funky with context, then redirect
			}
		}
	}
}
