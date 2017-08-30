package errorevent

import (
    "github.com/flimzy/goweb/event"
)

type ErrorEvent struct {
    *event.Event
    *js.Object
    // A string containing a human-readable error message describing the problem. Read only.
    Message string `js:"message"`
    // A string containing the name of the script file in which the error occurred. Read only.
    Filename string `js:"filename"`
    // An integer containing the line number of the script file on which the error occurred. Read only.
    LineNo int `js:"lineno"`
    // An integer containing the column number of the script file on which the error occurred. Read only.
    ColNo int `js:"colno"`
    // A JavaScript Object that is concerned by the event. Read only.
    Error *js.Object `js:"error"`
}
