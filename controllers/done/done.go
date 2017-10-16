// Package done provides a single card to display when there is nothing left to
// study.
package done

import (
	"context"
	"time"

	"github.com/FlashbackSRS/flashback"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
)

// Card is a card displaying a "no cards to study" message
type Card struct{}

var _ flashback.CardView = &Card{}

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
		studyview.ButtonRight: {
			Name:    "Check Again",
			Enabled: true,
		},
	}, nil
}

// Action always returns true, to allow checking for new due cards.
func (c *Card) Action(_ context.Context, _ *int, _ time.Time, _ interface{}) (done bool, err error) {
	return true, nil
}

//go:generate go-bindata -pkg done -nocompress -prefix files -o data.go files

// Body returns the Done card body.
func (c *Card) Body(_ context.Context, _ int) (string, error) {
	body, err := Asset("done.html")
	return string(body), err
}

// BuryRelated does nothing
func (c *Card) BuryRelated() error {
	return nil
}
