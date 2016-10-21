// +build js

package studyhandler

import (
	"encoding/base64"
	"net/url"

	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback/repository"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

var jQuery = jquery.NewJQuery

type cardState struct {
	Card *repo.Card
	Face int
}

var currentCard *cardState

// BeforeTransition prepares the page to study
func BeforeTransition(event *jquery.Event, ui *js.Object, p url.Values) bool {
	u, err := repo.CurrentUser()
	if err != nil {
		log.Printf("No user logged in: %s\n", err)
		return false
	}
	log.Debugf("card = %s\n", p.Get("card"))
	go func() {
		// Ensure the indexes are created before trying to use them
		u.DB()

		if currentCard != nil {
			if currentCard.Face++; currentCard.Face > 1 {
				currentCard = nil
			}
		}

		if currentCard == nil {
			card, err := u.GetNextCard()
			if err != nil {
				log.Printf("Error fetching card: %+v\n", err)
				return
			}
			currentCard = &cardState{
				Card: card,
				Face: 0,
			}
		}
		if err := DisplayFace(currentCard); err != nil {
			log.Printf("Erorr displaying card: %+v\n", err)
		}
	}()

	return true
}

func DisplayFace(c *cardState) error {
	body, iframeID, err := c.Card.Body(c.Face)
	if err != nil {
		return errors.Wrap(err, "Error fetching body")
	}
	responses, err := c.Card.Responses(c.Face)
	if err != nil {
		return errors.Wrap(err, "Error fetching responses")
	}

	doc := js.Global.Get("document")

	iframe := doc.Call("createElement", "iframe")
	iframe.Call("setAttribute", "sandbox", "allow-scripts")
	iframe.Call("setAttribute", "seamless", nil)
	iframe.Set("id", iframeID)
	iframe.Set("src", "data:text/html;charset=utf-8;base64,"+base64.StdEncoding.EncodeToString([]byte(body)))

	container := jQuery(":mobile-pagecontainer")

	jQuery("#cardframe", container).Append(iframe)

	r := jQuery("#responses", container)
	for _, response := range responses {
		li := doc.Call("createElement", "li")
		a := doc.Call("createElement", "a")
		a.Call("setAttribute", "data-role", "button")
		a.Call("setAttribute", "data-icon", response.Icon)
		a.Call("setAttribute", "data-lt", response.Name)
		a.Set("text", response.Display)
		li.Call("appendChild", a)
		r.Append(li)
	}
	r.Call("enhanceWithin")

	jQuery(".show-until-load", container).Hide()
	jQuery(".hide-until-load", container).Show()
	return nil
}
