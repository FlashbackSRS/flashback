// Package mock is the model handler for testing purposes
package mock

import (
	"fmt"
	"time"

	"github.com/flimzy/log"

	repo "github.com/FlashbackSRS/flashback/repository"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
)

// Mock is an Anki Basic model
type Mock struct {
	t string
}

var _ repo.ModelController = &Mock{}

// RegisterMock registers the mock Model as the requested type, for tests.
func RegisterMock(t string) {
	m := &Mock{t: t}
	repo.RegisterModelController(m)
}

// Type returns the string "anki-basic", to identify this model handler's type.
func (m *Mock) Type() string {
	return m.t
}

// IframeScript returns JavaScript to run inside the iframe.
func (m *Mock) IframeScript() []byte {
	return []byte(fmt.Sprintf(`
		/* Mock Model */
		console.log("Mock Model '%s'");
`, m.t))
}

// Buttons returns the initial buttons state
func (m *Mock) Buttons(_ int) (studyview.ButtonMap, error) {
	return studyview.ButtonMap{
		studyview.ButtonLeft: studyview.ButtonState{
			Name:    "Incorrect",
			Enabled: true,
		},
		studyview.ButtonCenterLeft: studyview.ButtonState{
			Name:    "Difficult",
			Enabled: true,
		},
		studyview.ButtonCenterRight: studyview.ButtonState{
			Name:    "Correct",
			Enabled: true,
		},
		studyview.ButtonRight: studyview.ButtonState{
			Name:    "Easy",
			Enabled: true,
		},
	}, nil
}

// Action responds to a card action, such as a button press
func (m *Mock) Action(card *repo.Card, face *int, _ time.Time, button studyview.Button) (bool, error) {
	log.Debugf("face: %d, button: %+v\n", face, button)
	return true, nil
}
