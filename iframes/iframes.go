// Package iframes provides abstractions for inter-iframe communication
package iframes

import (
	"fmt"

	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/pkg/errors"
)

// Message is a message to be sent to or received from an iframe
type Message struct {
	*js.Object
	Type     string     `js:"type"`
	IframeID string     `js:"iframeID"`
	CardID   string     `js:"cardID"`
	Payload  *js.Object `js:"payload"`
}

// Respond sends a response to the iframe
type Respond func(mtype string, payload interface{}, transferrable ...*js.Object) error

// Init installs the event listener to receive messages from iframes.
func Init() {
	js.Global.Call("addEventListener", "message", func(e *js.Object) {
		go func() {
			if err := receiveMessage(e); err != nil {
				log.Printf("Error processing message from iframe: %s\n", err)
			}
		}()
	})
}

// ListenerFunc is a function which responds to a parsed iframe message. It
// receives the original payload, and a function which may be called with a
// response to the original iframe.
//
// A ListenerFunc is called as a JS callback, so it may not block.
type ListenerFunc func(cardID string, payload *js.Object, respond Respond) error

var listeners = make(map[string]ListenerFunc)

// RegisterListener registers a ListnerFunc to receive messages of the given
// type.
func RegisterListener(mtype string, lf ListenerFunc) {
	if _, ok := listeners[mtype]; ok {
		panic("Listener for type '" + mtype + "' already registered")
	}
	listeners[mtype] = lf
}

func lookupCard(iframeID string) (string, error) {
	if cardID, ok := cardRegistry[iframeID]; ok {
		return cardID, nil
	}
	return "", errors.Errorf("iframe %s not registered", iframeID)
}

func receiveMessage(e *js.Object) error {
	msg := &Message{Object: e.Get("data")}
	cardID, err := lookupCard(msg.IframeID)
	if err != nil {
		return errors.Wrap(err, "invalid frame registration")
	}
	for mtype, lfunc := range listeners {
		if mtype == msg.Type {
			return lfunc(cardID, msg.Payload, responder(msg.IframeID))
		}
	}
	return errors.Errorf("unhandled message of type '%s' from iframe %s\n", msg.Type, msg.IframeID)
}

var cardRegistry = map[string]string{}

// RegisterIframe associates an iframe ID with a card ID so that requests can
// be served. A Responder is returned, which may then be used to send messages
// to the iframe.
func RegisterIframe(iframeID, cardID string) (Respond, error) {
	fmt.Printf("Registering %s for %s\n", iframeID, cardID)
	if card, ok := cardRegistry[iframeID]; ok {
		return nil, errors.Errorf("iframe already registered to card %s", card)
	}
	cardRegistry[iframeID] = cardID
	return responder(iframeID), nil
}

// UnregisterIframe unregisters the iframe. This should be called whenever
// the iframe is removed from the DOM. Any Respond functions corresponding
// to this iframe will be invalidated.
func UnregisterIframe(iframeID string) error {
	if _, ok := cardRegistry[iframeID]; !ok {
		return errors.New("iframe not registered")
	}
	delete(cardRegistry, iframeID)
	return nil
}

func responder(iframeID string) Respond {
	return Respond(func(mtype string, payload interface{}, transferable ...*js.Object) error {
		return sendMessage(mtype, iframeID, payload, transferable...)
	})
}

func sendMessage(mtype, iframeID string, payload interface{}, transferable ...*js.Object) error {
	log.Debugf("Sending response %s/%s/%v\n", mtype, iframeID, payload)
	if _, ok := cardRegistry[iframeID]; !ok {
		return errors.Errorf("iframe '%s' no registered", iframeID)
	}
	var iframe *js.Object
	iframes := js.Global.Get("document").Call("getElementsByTagName", "iframe")
	for i := 0; i < iframes.Length(); i++ {
		if iframes.Index(i).Get("src").String() == iframeID {
			iframe = iframes.Index(i)
			break
		}
	}
	if iframe == js.Undefined {
		return errors.Errorf("iframe not found in DOM")
	}
	iframe.Get("contentWindow").Call("postMessage", map[string]interface{}{
		"type":    mtype,
		"payload": payload,
	}, "*", transferable)
	return nil
}
