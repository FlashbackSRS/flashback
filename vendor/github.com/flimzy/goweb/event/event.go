package event

import (
	"fmt"
	"time"

	"github.com/gopherjs/gopherjs/js"
)

var monotonic bool
var startTime = time.Now()

func init() {
	if st, err := navStart(); err == nil {
		startTime = time.Unix(st/1000, (st%1000)*1000000)
	}
	now := New("x", Init{}).Get("timeStamp").Int64()
	// If we add the event's timestamp to our start time, and we get a result after time.Now(),
	// it means the event's timestamp was in absolute seconds, so we don't support monotonic
	// timestamps. We pad time.Now() by one minute, to account for any start up time.
	monotonic = !startTime.Add(time.Duration(now) * time.Millisecond).After(time.Now().Add(1 * time.Minute))
}

func navStart() (t int64, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(*js.Error)
		}
	}()
	t = js.Global.Get("window").Get("performance").Get("timing").Get("navigationStart").Int64()
	return
}

// Event represents any event of the DOM. It contains common properties and methods to any event.
//
// Many other objects extend the Event object through embedding.
//
// See https://developer.mozilla.org/en-US/docs/Web/API/Event
type Event struct {
	*js.Object
	// A boolean indicating whether the event bubbles up through the DOM or not. Read only.
	Bubbles bool `js:"bubbles"`
	// A boolean indicating whether the event is cancelable. Read only.
	Cancelable bool `js:"cancelable"`
	// A reference to the currently registered target for the event. Read only.
	CurrentTarget *js.Object `js:"currentTarget"`
	// Indicates whether or not event.preventDefault() has been called on the event. Read only.
	DefaultPrevented bool `js:"defaultPrevented"`
	// Indicates which phase of the event flow is being processed. Read only.
	EventPhase int `js:"eventPhase"`
	// A reference to the target to which the event was originally dispatched. Read only.
	Target *js.Object `js:"target"`
	// The name of the event (case-insensitive). Read only.
	Type string `js:"type"`
	// Indicates whether or not the event was initiated by the browser (after a user click for instance) or by a script (using an event creation method, like event.initEvent). Read only.
	IsTrusted bool `js:"isTrusted"`
}

type Init struct {
	Bubbles    bool `js:"bubbles"`
	Cancelable bool `js:"cancelable"`
}

func New(name string, init Init) *Event {
	o := js.Global.Get("Event").New(name, init)
	return &Event{Object: o}
}

func Internalize(o *js.Object) *Event {
	return &Event{Object: o}
}

func (e *Event) PreventDefault() {
	e.Call("preventDefault")
}

func (e *Event) StopImmediatePropagation() {
	e.Call("stopImmediatePropagation")
}

func (e *Event) StopPropagation() {
	e.Call("stopPropagation")
}

// The time that the event was created.
func (e *Event) Timestamp() time.Time {
	ts := e.Get("timeStamp").Int64()
	if monotonic {
		return startTime.Add(time.Duration(ts) * time.Millisecond)
	}
	return time.Unix(ts/1000, (ts%1000)*1000000)
}

type EventHandler interface {
	HandleEvent(*Event)
}

type EventHandlerFunc func(*Event)

type ListenerOpts struct {
	*js.Object
	Capture bool `js:"capture"`
	Passive bool `js:"passive"`
}

func AddEventListener(target interface{}, name string, handler EventHandlerFunc, opts ListenerOpts) {
	o := target.(*js.Object)
	fmt.Printf("adding event listener %s\n", name)
	o.Call("addEventListener", name, func(e *js.Object) {
		handler(Internalize(e))
	}, opts)
}

// type ErrorEvent struct {
//     *Event
//     *js.Object
//     // A string containing a human-readable error message describing the problem. Read only.
//     Message string `js:"message"`
//     // A string containing the name of the script file in which the error occurred. Read only.
//     Filename string `js:"filename"`
//     // An integer containing the line number of the script file on which the error occurred. Read only.
//     LineNo int `js:"lineno"`
//     // An integer containing the column number of the script file on which the error occurred. Read only.
//     ColNo int `js:"colno"`
//     // A JavaScript Object that is concerned by the event. Read only.
//     Error *js.Object `js:"error"`
// }
