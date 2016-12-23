// Package mock is the model handler for testing purposes
package mock

import (
	"fmt"

	"github.com/FlashbackSRS/flashback/cardmodel"
)

// Model is an Anki Basic model
type Model struct {
	t string
}

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
func (m *Model) Buttons(_ int) cardmodel.AnswerButtonsState {
	return cardmodel.AnswerButtonsState{
		cardmodel.AnswerButton{
			Name:    "Wrong Answer",
			Enabled: true,
		},
		cardmodel.AnswerButton{
			Name:    "Correct Answer",
			Enabled: true,
		},
		cardmodel.AnswerButton{
			Name:    "Easy Answer",
			Enabled: true,
		},
	}
}
