package jqeventrouter

import (
	"fmt"
	"net/url"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/console"
)

type EventMux struct {
	paths   map[string]*muxEntry
	uriFunc func(*jquery.Event, *js.Object) string
}

type muxEntry struct {
	h    Handler
	path string
}

func NewEventMux() *EventMux {
	return &EventMux{paths: make(map[string]*muxEntry)}
}

func (mux *EventMux) HandleEvent(event *jquery.Event, data *js.Object, _ url.Values) bool {
	rawUri := mux.getUri(event, data)
	reqUri, err := url.ParseRequestURI(rawUri)
	if err != nil {
		fmt.Printf("Error parsing path '%s': %s\n", rawUri, err)
		return true
	}
	console.Log("URI = %s", rawUri)
	for path, entry := range mux.paths {
		if reqUri.Path == path {
			// We found a match!
			return entry.h.HandleEvent(event, data, reqUri.Query())
		}
	}
	return true
}

// SetUriFunc allows you to specify a custom function to determine the URI of a request.
// The custom function is expected to return a string, representing the URI.
// If unset, everything after the hostname in window.location.href is retunred.
func (mux *EventMux) SetUriFunc(fn func(event *jquery.Event, data *js.Object) string) {
	mux.uriFunc = fn
}

func (mux *EventMux) getUri(event *jquery.Event, data *js.Object) string {
	if mux.uriFunc != nil {
		return mux.uriFunc(event, data)
	}
	return js.Global.Get("location").Get("href").String()
}

// Handle registers the handler for the given path. If a handler already
// exists for path, or the path is not valid in a URL, Handle panics.
func (mux *EventMux) Handle(path string, handler Handler) {
	if path == "" {
		panic("jqeventrouter: empty path")
	}
	if handler == nil {
		panic("jqeventrouter: nil handler")
	}
	if parsed, err := url.ParseRequestURI(path); err != nil {
		panic(fmt.Sprintf("jqeventrouter: Invalid path '%s': %s", path, err))
	} else if parsed.Path != path {
		panic(fmt.Sprintf("jqeventrouter: Invalid path '%s'. Did you mean '%s'?", path, parsed.Path))
	}
	if _, ok := mux.paths[path]; ok {
		panic("jqeventrouter: multiple registrations for " + path)
	}

	mux.paths[path] = &muxEntry{
		h:    handler,
		path: path,
	}
}

func (mux *EventMux) HandleFunc(path string, handler func(*jquery.Event, *js.Object, url.Values) bool) {
	mux.Handle(path, HandlerFunc(handler))
}

// Listen is a convenience function which calls Listen(event,mux)
func (mux *EventMux) Listen(event string) {
	Listen(event, mux)
}

type Handler interface {
	HandleEvent(event *jquery.Event, data *js.Object, params url.Values) bool
}

// The HandlerFunc type is an adaptor to allow the use of ordinary functions
// as Event handlers. If f is a function of the appropriate signature, HandlerFunc(f)
// is a Handler object that calls f.
type HandlerFunc func(*jquery.Event, *js.Object, url.Values) bool

// HandleEvent calls f(this,event)
func (f HandlerFunc) HandleEvent(event *jquery.Event, data *js.Object, p url.Values) bool {
	return f(event, data, p)
}

type EventListener struct {
	event    string
	listener func(*jquery.Event, *js.Object) bool
	detached bool
}

// Listen attaches the Handler to the window and begins listening for the specified
// jQuery event, reterning an EventListener object
func Listen(event string, handler Handler) *EventListener {
	console.Log("Adding jQuery event listener")
	listener := func(event *jquery.Event, data *js.Object) bool {
		return handler.HandleEvent(event, data, url.Values{})
	}
	jquery.NewJQuery(js.Global.Get("document")).On(event, listener)
	return &EventListener{event: event, listener: listener}
}

// UnListen detaches the EventListener from the window.
func (l *EventListener) UnListen() {
	if l.detached == true {
		panic("Already detached")
	}
	jquery.NewJQuery(js.Global.Get("document")).Off(l.event, l.listener)
	l.detached = true
}

// NullHandler returns a null Handler object, which unconditionally returns true.
// It can be used for terminating event chains that don't need to affect the page.
func NullHandler() Handler {
	return HandlerFunc(func(_ *jquery.Event, _ *js.Object, _ url.Values) bool {
		return true
	})
}
