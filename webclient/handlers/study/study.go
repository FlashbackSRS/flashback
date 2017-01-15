// +build js

package studyhandler

import (
	"net/url"
	"time"

	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback/cardmodel"
	"github.com/FlashbackSRS/flashback/fserve"
	"github.com/FlashbackSRS/flashback/repository"
)

var jQuery = jquery.NewJQuery

type cardState struct {
	Card      *repo.Card
	StartTime time.Time
	Face      int
}

var currentCard *cardState

// BeforeTransition prepares the page to study
func BeforeTransition(event *jquery.Event, ui *js.Object, _ url.Values) bool {
	u, err := repo.CurrentUser()
	if err != nil {
		log.Printf("No user logged in: %s\n", err)
		return false
	}
	go func() {
		if err := ShowCard(u); err != nil {
			log.Printf("Error showing card: %+v", err)
		}
	}()

	return true
}

func ShowCard(u *repo.User) error {
	// Ensure the indexes are created before trying to use them
	u.DB()

	if currentCard == nil {
		card, err := u.GetNextCard()
		if err != nil {
			return errors.Wrap(err, "fetch card")
		}
		if card == nil {
			return errors.New("got a nil card")
		}
		currentCard = &cardState{
			Card: card,
		}
	}
	log.Debugf("Card ID: %s\n", currentCard.Card.DocID())

	mh, err := currentCard.Card.ModelHandler()
	if err != nil {
		return errors.Wrap(err, "failed to get card's model handler")
	}

	log.Debug("Setting up the buttons\n")
	buttons := jQuery(":mobile-pagecontainer").Find("#answer-buttons").Find(`[data-role="button"]`)
	buttons.On("click", func(e *js.Object) {
		buttons.Off() // Make sure we don't accept other press events
		id := e.Get("currentTarget").Call("getAttribute", "data-id").String()
		log.Debugf("Button %s was pressed!\n", id)
		HandleCardAction(cardmodel.Button(id))
	})
	buttonAttrs, err := mh.Buttons(currentCard.Face)
	if err != nil {
		return errors.Wrap(err, "failed to get buttons list")
	}
	for i := 0; i < buttons.Length; i++ {
		button := jQuery(buttons.Underlying().Index(i))
		id := button.Attr("data-id")
		attr, ok := buttonAttrs[(cardmodel.Button(id))]
		button.Call("button")
		if !ok {
			button.SetText(" ")
		} else {
			button.SetText(attr.Name)
			if attr.Enabled {
				button.Call("button", "enable")
			} else {
				button.Call("button", "disable")
			}
		}
	}

	body, iframeID, err := currentCard.Card.Body(currentCard.Face)
	fserve.RegisterIframe(iframeID, currentCard.Card.DocID())
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

	jQuery("#cardframe", container).Empty().Append(iframe)

	jQuery(".show-until-load", container).Hide()
	jQuery(".hide-until-load", container).Show()
	currentCard.StartTime = time.Now()
	return nil
}

func HandleCardAction(button cardmodel.Button) {
	card := currentCard.Card
	mh, err := card.ModelHandler()
	if err != nil {
		log.Printf("failed to get card's model handler: %s\n", err)
	}
	face := currentCard.Face
	done, err := mh.Action(card.Card, &currentCard.Face, currentCard.StartTime, cardmodel.Action{
		Button: button,
	})
	if err != nil {
		log.Printf("Error executing card action for face %d / %+v: %s", face, card, err)
	}
	if done {
		currentCard = nil
	} else {
		if face == currentCard.Face {
			log.Printf("face wasn't incremented!\n")
		}
	}
	jQuery(":mobile-pagecontainer").Call("pagecontainer", "change", "/study.html")
}
