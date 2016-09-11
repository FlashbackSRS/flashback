// +build js

package studyhandler

import (
	// 	"bytes"
	// 	"errors"
	"fmt"
	// 	"html/template"
	"net/url"
	// 	"strings"

	// 	"github.com/pborman/uuid"

	// 	"github.com/flimzy/go-pouchdb"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	// 	"github.com/gopherjs/jsbuiltin"
	// 	"golang.org/x/net/html"

	// 	"github.com/FlashbackSRS/flashback/util"
	"github.com/FlashbackSRS/flashback/repository"
)

var jQuery = jquery.NewJQuery

// BeforeTransition prepares the page to study
func BeforeTransition(event *jquery.Event, ui *js.Object, p url.Values) bool {
	u, err := repo.CurrentUser()
	if err != nil {
		fmt.Printf("No user logged in: %s\n", err)
		return false
	}
	fmt.Printf("card = %s\n", p.Get("card"))
	go func() {
		container := jQuery(":mobile-pagecontainer")
		// Ensure the indexes are created before trying to use them
		u.DB()

		card, err := repo.GetCard()
		if err != nil {
			fmt.Printf("Error fetching card: %+v\n", err)
			return
		}
		fmt.Printf("card = %s\n", card)
		body, iframeID, err := card.Body()
		if err != nil {
			fmt.Printf("Error parsing body: %+v\n", err)
		}
		fmt.Printf("body = %s\niframe = %s\n", body, iframeID)
		// 		body, iframeId, err := getCardBodies(card)
		// 		if err != nil {
		// 			fmt.Printf("Error reading card: %s\n", err)
		// 		}

		// 		iframe := js.Global.Get("document").Call("createElement", "iframe")
		// 		iframe.Call("setAttribute", "sandbox", "allow-scripts")
		// 		iframe.Call("setAttribute", "seamless", nil)
		// 		iframe.Set("id", iframeId)
		// 		iframe.Set("src", "data:text/html;charset=utf-8,"+jsbuiltin.EncodeURI(body))

		// 		js.Global.Get("document").Call("getElementById", "cardframe").Call("appendChild", iframe)

		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()
	}()

	return true
}
