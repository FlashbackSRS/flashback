package worker

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/flimzy/goweb/event"
)

type Worker struct {
	js.Object
	OnError event.EventHandlerFunc `js:"onerror"`
	OnMessage event.EventHandlerFunc `js:"onmessage"`
}

func New(url string) (w *Worker, e error) {
	defer func() {
		if r := recover(); r != nil {
			e = r.(*js.Error)
		}
	}()
	o := js.Global.Get("Worker").New(url)
	w = &Worker{Object: *o}
	return
}

func (w *Worker) PostMessage(msg interface{}) {
	w.Call("postMessage", msg)
}

func (w *Worker) Terminate() {
	w.Call("terminate")
}
