// Package done implements a controller for the 'Nothing to study now' page.
package done

import (
	"time"

	"github.com/FlashbackSRS/flashback-model"
	repo "github.com/FlashbackSRS/flashback/repository"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
)

func init() {
	repo.RegisterModelController(&Done{})
}

// Done is the done-stydying controller
type Done struct{}

var _ repo.ModelController = &Done{}

// Type returns the string "done"
func (d *Done) Type() string {
	return "done"
}

// IframeScript returns nothing for the Done controller.
func (d *Done) IframeScript() []byte {
	return nil
}

// Buttons returns the list of available buttons. As there is only one face
// for Done cards, the face value is ignored.
func (d *Done) Buttons(_ int) (studyview.ButtonMap, error) {
	return studyview.ButtonMap{
		studyview.ButtonRight: studyview.ButtonState{
			Name:    "Check Again",
			Enabled: true,
		},
	}, nil
}

// Action always returns true, to allow checking for new due cards.
func (d *Done) Action(_ *repo.Card, _ *int, _ time.Time, _ studyview.Button) (done bool, err error) {
	return true, nil
}

func Card() *fb.Card {

}
