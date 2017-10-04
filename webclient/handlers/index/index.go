// +build js

package index

import (
	"bytes"
	"context"
	"net/url"

	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/model"
	"github.com/flimzy/jqeventrouter"
)

var jQuery = jquery.NewJQuery

// BeforeTransition loads the index for display.
func BeforeTransition(repo *model.Repo) jqeventrouter.HandlerFunc {
	return func(_ *jquery.Event, ui *js.Object, _ url.Values) bool {
		log.Debugf("index handler")
		container := jQuery(ui.Get("toPage"))
		// container := jQuery(":mobile-pagecontainer")

		tmpl, err := deckListTemplate()
		if err != nil {
			log.Printf("Failed to read/parse desk list template: %s", err)
			return false
		}

		go func() {
			decks, err := repo.DeckList(context.TODO())
			if err != nil {
				log.Printf("Failed to read deck list: %s", err)
			}

			buf := &bytes.Buffer{}
			if err := tmpl.Execute(buf, decks); err != nil {
				log.Printf("Failed to execute template: %s", err)
				return
			}

			jQuery("#deck-list", container).SetHtml(buf.String())
			jQuery(".show-until-load", container).Hide()
			jQuery(".hide-until-load", container).Show()

		}()
		return true
	}
}
