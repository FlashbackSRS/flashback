// +build js

package studyhandler

import (
	"fmt"
	"net/url"

	"honnef.co/go/js/console"

	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback/fserve"
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
		if err := ShowCard(u); err != nil {
			log.Printf("Error showing card: %+v", err)
		}
	}()

	return true
}

func ShowCard(u *repo.User) error {
	log.Debug("ShowCard()\n")
	// Ensure the indexes are created before trying to use them
	u.DB()

	if currentCard == nil {
		card, err := u.GetNextCard()
		if err != nil {
			return errors.Wrap(err, "fetch card")
		}
		currentCard = &cardState{
			Card: card,
			Face: 0,
		}
	}
	log.Debugf("Card ID: %s\n", currentCard.Card.DocID())

	mh, err := currentCard.Card.ModelHandler()
	if err != nil {
		return err
	}
	log.Debug("Got the model handler\n")

	buttons := jQuery(":mobile-pagecontainer").Find("#answer-buttons").Find(`[data-role="button"]`)
	console.Log(buttons)
	log.Debug("Setting up the buttons\n")
	for i, b := range mh.Buttons(currentCard.Face) {
		log.Debugf("Setting button %d to %s\n", i, b.Name)
		button := jQuery(buttons.Underlying().Index(i))
		if b.Name == "" {
			// Hack to enforce same-height buttons
			button.SetText(" ")
		} else {
			button.SetText(b.Name)
		}
		button.Call("button")
		if b.Enabled {
			button.Call("button", "enable")
		} else {
			button.Call("button", "disable")
		}
	}
	buttons.On("click", ButtonPressed)

	if err := DisplayFace(currentCard); err != nil {
		return errors.Wrap(err, "display card")
	}
	return nil
}

func ButtonPressed(e *js.Object) {
	fmt.Printf("Button %s was pressed!\n", e.Get("currentTarget").Call("getAttribute", "data-id").String())
}

func DisplayFace(c *cardState) error {
	body, iframeID, err := c.Card.Body(c.Face)
	fserve.RegisterIframe(iframeID, c.Card.DocID())
	if err != nil {
		return errors.Wrap(err, "Error fetching body")
	}

	doc := js.Global.Get("document")

	iframe := doc.Call("createElement", "iframe")
	iframe.Call("setAttribute", "sandbox", "allow-scripts")
	iframe.Call("setAttribute", "seamless", nil)
	iframe.Set("id", iframeID)
	ab := js.NewArrayBuffer([]byte(body))
	b := js.Global.Get("Blob").New([]interface{}{ab}, map[string]string{"type": "text/html"})
	iframe.Set("src", js.Global.Get("URL").Call("createObjectURL", b))

	container := jQuery(":mobile-pagecontainer")

	jQuery("#cardframe", container).Append(iframe)

	jQuery(".show-until-load", container).Hide()
	jQuery(".hide-until-load", container).Show()
	return nil
}
