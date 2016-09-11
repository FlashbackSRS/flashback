// +build js

package studyhandler

import (
	"encoding/base64"
	"net/url"

	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/repository"
)

var jQuery = jquery.NewJQuery

// BeforeTransition prepares the page to study
func BeforeTransition(event *jquery.Event, ui *js.Object, p url.Values) bool {
	u, err := repo.CurrentUser()
	if err != nil {
		log.Printf("No user logged in: %s\n", err)
		return false
	}
	log.Debugf("card = %s\n", p.Get("card"))
	go func() {
		container := jQuery(":mobile-pagecontainer")
		// Ensure the indexes are created before trying to use them
		u.DB()

		card, err := repo.GetRandomCard()
		if err != nil {
			log.Printf("Error fetching card: %+v\n", err)
			return
		}
		log.Debugf("card = %s\n", card)
		body, iframeID, err := card.Body()
		if err != nil {
			log.Printf("Error parsing body: %+v\n", err)
			return
		}
		log.Debugf("body = %s\niframe = %s\n", body, iframeID)

		iframe := js.Global.Get("document").Call("createElement", "iframe")
		iframe.Call("setAttribute", "sandbox", "allow-scripts")
		iframe.Call("setAttribute", "seamless", nil)
		iframe.Set("id", iframeID)
		iframe.Set("src", "data:text/html;charset=utf-8;base64,"+base64.StdEncoding.EncodeToString([]byte(body)))

		js.Global.Get("document").Call("getElementById", "cardframe").Call("appendChild", iframe)

		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()
	}()

	return true
}
