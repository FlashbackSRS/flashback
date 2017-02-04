// +build js

package studyhandler

import (
	"net/url"
	"time"

	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback/iframes"
	"github.com/FlashbackSRS/flashback/repository"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
)

var jQuery = jquery.NewJQuery

type cardState struct {
	Card      repo.Card
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

	body, err := currentCard.Card.Body(currentCard.Face)
	if err != nil {
		return errors.Wrap(err, "fetching body")
	}

	iframe := js.Global.Get("document").Call("createElement", "iframe")
	iframe.Call("setAttribute", "sandbox", "allow-scripts")
	iframe.Call("setAttribute", "seamless", nil)
	ab := js.NewArrayBuffer([]byte(body))
	b := js.Global.Get("Blob").New([]interface{}{ab}, map[string]string{"type": "text/html"})
	iframeURL := js.Global.Get("URL").Call("createObjectURL", b)
	iframe.Set("src", iframeURL)
	respond, err := iframes.RegisterIframe(iframeURL.String(), currentCard.Card.DocID())
	if err != nil {
		return errors.Wrap(err, "failed to register iframe")
	}

	log.Debug("Setting up the buttons\n")
	buttons := jQuery(":mobile-pagecontainer").Find("#answer-buttons").Find(`[data-role="button"]`)
	buttons.RemoveClass("ui-btn-active")
	clickFunc := func(e *js.Object) {
		go func() { // DB updates block
			buttons.Off() // Make sure we don't accept other press events
			buttonID := e.Get("currentTarget").Call("getAttribute", "data-id").String()
			log.Debugf("Button %s was pressed!\n", buttonID)
			respond("submit", buttonID)
		}()
	}
	buttonAttrs, err := currentCard.Card.Buttons(currentCard.Face)
	if err != nil {
		return errors.Wrap(err, "failed to get buttons list")
	}
	for i := 0; i < buttons.Length; i++ {
		button := jQuery(buttons.Underlying().Index(i))
		id := button.Attr("data-id")
		button.Call("button")
		attr, _ := buttonAttrs[(studyview.Button(id))] // I can ignore the ok value, because the nil value for attr works the same
		name := attr.Name
		if name == "" {
			name = " "
		}
		button.SetText(name)
		if attr.Enabled {
			button.Call("button", "enable")
			button.On("click", clickFunc)
		} else {
			button.Call("button", "disable")
		}
	}

	container := jQuery(":mobile-pagecontainer")

	oldIframes := jQuery("#cardframe", container).Find("iframe").Underlying()
	for i := 0; i < oldIframes.Length(); i++ {
		oldIframeID := oldIframes.Index(i).Get("src").String()
		if err := iframes.UnregisterIframe(oldIframeID); err != nil {
			log.Printf("Failed to unregister old iframe '%s': %s\n", oldIframeID, err)
		}
		js.Global.Get("URL").Call("revokeObjectURL", oldIframeID)
	}

	jQuery("#cardframe", container).Empty().Append(iframe)

	jQuery(".show-until-load", container).Hide()
	jQuery(".hide-until-load", container).Show()
	currentCard.StartTime = time.Now()
	return nil
}

func init() {
	iframes.RegisterListener("submit", handleSubmit)
}

func handleSubmit(cardID string, payload *js.Object, _ iframes.Respond) error {
	card := currentCard.Card
	face := currentCard.Face
	if card.DocID() != cardID {
		return errors.Errorf("received submit for unexpected card. Got %s, expected %s", cardID, card.DocID())
	}
	done, err := card.Action(&currentCard.Face, currentCard.StartTime, payload)
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
	return nil
}
