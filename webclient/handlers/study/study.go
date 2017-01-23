// +build js

package studyhandler

import (
	"math/rand"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback/fserve"
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

	log.Debug("Setting up the buttons\n")
	buttons := jQuery(":mobile-pagecontainer").Find("#answer-buttons").Find(`[data-role="button"]`)
	buttons.RemoveClass("ui-btn-active")
	clickFunc := func(e *js.Object) {
		go func() { // DB updates block
			buttons.Off() // Make sure we don't accept other press events
			id := e.Get("currentTarget").Call("getAttribute", "data-id").String()
			log.Debugf("Button %s was pressed!\n", id)
			HandleCardAction(studyview.Button(id))
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

	body, err := currentCard.Card.Body(currentCard.Face)
	if err != nil {
		return errors.Wrap(err, "fetching body")
	}
	iframeID := GenerateIframeID(8)
	body, err = addIframeID(body, iframeID)
	if err != nil {
		return errors.Wrap(err, "failed to add iframe metadata")
	}
	fserve.RegisterIframe(iframeID, currentCard.Card.DocID())

	iframe := js.Global.Get("document").Call("createElement", "iframe")
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

func addIframeID(body, iframeID string) (newBody string, err error) {
	r := strings.NewReader(body)
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse body")
	}
	doc.Find("head").AppendHtml(`<meta name="iframeid" content="` + iframeID + `">`)
	return doc.Html()
}

func HandleCardAction(button studyview.Button) {
	card := currentCard.Card
	face := currentCard.Face
	done, err := card.Action(&currentCard.Face, currentCard.StartTime, button)
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

// Random number function borrowed from http://stackoverflow.com/a/31832326/13860
var src = rand.NewSource(time.Now().UnixNano())

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// GenerateIframeID returns a random string of n characters, with 6 random bits
// per character.
func GenerateIframeID(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}
