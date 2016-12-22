// Package mock is the model handler for testing purposes
package mock

import "github.com/FlashbackSRS/flashback/models"

// Model is an Anki Basic model
type Model struct{}

// Type returns the string "anki-basic", to identify this model handler's type.
func (m *Model) Type() string {
	return "mock-model"
}

// IframeScript returns JavaScript to run inside the iframe.
func (m *Model) IframeScript() []byte {
	return []byte(`
        /* Placeholder JS */
        console.log("Mock Handler");
    `)
}

// Buttons returns the initial buttons state
func (m *Model) Buttons(_ int) models.AnswerButtonsState {
	return AnswerButtonsState{
		models.AnswerButton{
			Name:    "Wrong Answer",
			Enabled: true,
		},
		models.AnswerButton{
			Name:    "Correct Answer",
			Enabled: true,
		},
		models.AnswerButton{
			Name:    "Easy Answer",
			Enabled: true,
		},
	}
}
