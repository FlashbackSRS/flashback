package fserve

import (
	"sync"

	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback/repository"
)

// Request represents an fserve request, sent from a card frame
type Request struct {
	IframeID string
	Tag      string
	CardID   string
	Path     string
}

// ParseRequest converts the raw request payload into a Request struct.
func ParseRequest(o *js.Object) (*Request, error) {
	data := o.Get("data")
	values := make(map[string]string)
	for _, key := range []string{"IframeID", "Tag", "CardID", "Path"} {
		val := data.Get(key)
		if jsbuiltin.TypeOf(o) == "undefined" {
			return nil, errors.Errorf("request missing key %s", key)
		}
		if val.String() == "" {
			return nil, errors.Errorf("no value for key %s", key)
		}
		values[key] = val.String()
	}
	return &Request{
		IframeID: values["IframeID"],
		Tag:      values["Tag"],
		CardID:   values["CardID"],
		Path:     values["Path"],
	}, nil
}

// Response represents a file serve response, sent back to the card frame.
type Response struct {
	Tag         string
	Path        string
	ContentType string
	Data        *js.Object
}

// Init installs the fserve as an event listener, listening for messages
func Init(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		js.Global.Call("addEventListener", "message", func(e *js.Object) {
			go func() { // So we can block if we want to
				if err := fserve(e); err != nil {
					log.Printf("Error serving file: %s", err)
				}
			}()
		})
	}()
}

var cardRegistry = map[string]string{}

// RegisterIframe associates an iframe ID with a card ID so that files can
// be served to it. Without this, any file requests will be refused, as a
// security precaution.
func RegisterIframe(iframeID, cardID string) error {
	if card, ok := cardRegistry[iframeID]; ok {
		return errors.Errorf("iframe already registered to card %s", card)
	}
	cardRegistry[iframeID] = cardID
	return nil
}

// UnregisterIframe unregisters the iframe. This should be called whenever
// the iframe is removed from the DOM.
func UnregisterIframe(iframeID string) error {
	if _, ok := cardRegistry[iframeID]; !ok {
		return errors.New("iframe not registered")
	}
	delete(cardRegistry, iframeID)
	return nil
}

func fserve(e *js.Object) error {
	req, err := ParseRequest(e)
	if err != nil {
		return errors.Wrap(err, "invalid request")
	}
	// var req Request
	// if err := json.Unmarshal([]byte(e.Get("data").String()), &req); err != nil {
	// 	return errors.Wrap(err, "decode request")
	// }
	if permittedCard, ok := cardRegistry[req.IframeID]; ok {
		if permittedCard != req.CardID {
			return errors.Errorf("iframe %s not registered for card %s", req.IframeID, req.CardID)
		}
	} else {
		return errors.Errorf("iframe %s not registered", req.IframeID)
	}
	log.Debugf("Request from iframe %s for card %s authorized for '%s'", req.IframeID, req.CardID, req.Path)
	att, err := fetchAttachment(req.CardID, req.Path)
	if err != nil {
		return errors.Wrap(err, "fetch file")
	}
	return errors.Wrap(sendResponse(req.IframeID, req.Tag, req.Path, att), "send response")
}

func fetchAttachment(cardID, filename string) (*repo.Attachment, error) {
	log.Debugf("Attempting to fetch '%s' for %s\n", filename, cardID)
	u, err := repo.CurrentUser()
	if err != nil {
		return nil, errors.Wrap(err, "current user")
	}
	card, err := u.GetCard(cardID)
	if err != nil {
		return nil, errors.Wrap(err, "fetch card")
	}
	return card.GetAttachment(filename)
}

func sendResponse(iframeID, tag, filename string, att *repo.Attachment) error {
	iframe := js.Global.Get("document").Call("getElementById", iframeID)
	if jsbuiltin.TypeOf(iframe) == "undefined" {
		return errors.Errorf("iframe not found in DOM")
	}
	ab := js.NewArrayBuffer(att.Content)
	log.Debugf("Before send, ab has %d bytes\n", ab.Get("byteLength").Int())
	iframe.Get("contentWindow").Call("postMessage", Response{
		Tag:         tag,
		Path:        filename,
		ContentType: att.ContentType,
		Data:        ab,
	}, "*", []interface{}{ab})
	log.Debugf("After send, ab has %d bytes\n", ab.Get("byteLength").Int())
	return nil
}
