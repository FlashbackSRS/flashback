// Package mock is the model handler for testing purposes
package mock

import (
	"fmt"

	"github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/cardmodel"
	"github.com/flimzy/log"
)

// Model is an Anki Basic model
type Model struct {
	t string
}

var _ cardmodel.Model = &Model{}

// RegisterMock registers the mock Model as the requested type, for tests.
func RegisterMock(t string) {
	m := &Model{t: t}
	cardmodel.RegisterModel(m)
}

// Type returns the string "anki-basic", to identify this model handler's type.
func (m *Model) Type() string {
	return m.t
}

// IframeScript returns JavaScript to run inside the iframe.
func (m *Model) IframeScript() []byte {
	return []byte(fmt.Sprintf(`
		/* Mock Model */
		console.log("Mock Model '%s'");
`, m.t))
}

// Buttons returns the initial buttons state
func (m *Model) Buttons(_ int) (*cardmodel.ButtonMap, error) {
	return &cardmodel.ButtonMap{
		cardmodel.ButtonLeft: cardmodel.AnswerButton{
			Name:    "Incorrect",
			Enabled: true,
		},
		cardmodel.ButtonCenterLeft: cardmodel.AnswerButton{
			Name:    "Difficult",
			Enabled: true,
		},
		cardmodel.ButtonCenterRight: cardmodel.AnswerButton{
			Name:    "Correct",
			Enabled: true,
		},
		cardmodel.ButtonRight: cardmodel.AnswerButton{
			Name:    "Easy",
			Enabled: true,
		},
	}, nil
}

// Action responds to a card action, such as a button press
func (m *Model) Action(card *fb.Card, face *int, action cardmodel.Action) (bool, error) {
	log.Debugf("face: %d, action: %+v\n", face, action)
	return true, nil
}
