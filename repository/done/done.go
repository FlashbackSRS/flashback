// Package done provides a single card to display when there is nothing left to
// study.
package done

import (
	"time"

	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
	"github.com/gopherjs/gopherjs/js"
)

// Card is a card displaying a "no cards to study" message
type Card struct{}

// GetCard returns the done card.
func GetCard() *Card {
	return &Card{}
}

// DocID returns a dummy document ID.
func (c *Card) DocID() string {
	return "<<done>>"
}

// Buttons returns the list of available buttons. As there is only one face
// for Done cards, the face value is ignored.
func (c *Card) Buttons(_ int) (studyview.ButtonMap, error) {
	return studyview.ButtonMap{
		studyview.ButtonRight: studyview.ButtonState{
			Name:    "Check Again",
			Enabled: true,
		},
	}, nil
}

// Action always returns true, to allow checking for new due cards.
func (c *Card) Action(_ *int, _ time.Time, _ *js.Object) (done bool, err error) {
	return true, nil
}

// Body returns the Done card body.
func (c *Card) Body(_ int) (string, error) {
	body, err := Asset("done.html")
	return string(body), err
}
