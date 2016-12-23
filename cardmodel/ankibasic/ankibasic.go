// Package ankibasic is the model handler for the Basic Anki model type.
package ankibasic

import "github.com/FlashbackSRS/flashback/cardmodel"

const (
	// FaceQuestion is the question face of a card
	FaceQuestion = iota
	// FaceAnswer is the answer face of a card
	FaceAnswer
)

// Model is an Anki Basic model
type Model struct{}

func init() {
	cardmodel.RegisterModel(&Model{})
}

// Type returns the string "anki-basic", to identify this model handler's type.
func (m *Model) Type() string {
	return "anki-basic"
}

// IframeScript returns JavaScript to run inside the iframe.
func (m *Model) IframeScript() []byte {
	data, err := Asset("script.js")
	if err != nil {
		panic(err)
	}
	return data
}

// Buttons returns the initial button state
func (m *Model) Buttons(face int) cardmodel.AnswerButtonsState {
	switch face {
	case FaceQuestion:
		return cardmodel.AnswerButtonsState{
			cardmodel.AnswerButton{
				Name:    "",
				Enabled: false,
			},
			cardmodel.AnswerButton{
				Name:    "",
				Enabled: false,
			},
			cardmodel.AnswerButton{
				Name:    "Show Answer",
				Enabled: true,
			},
		}
	case FaceAnswer:
		return cardmodel.AnswerButtonsState{
			cardmodel.AnswerButton{
				Name:    "Wrong Answer",
				Enabled: false,
			},
			cardmodel.AnswerButton{
				Name:    "Mostly Correct",
				Enabled: false,
			},
			cardmodel.AnswerButton{
				Name:    "Correct Answer",
				Enabled: true,
			},
		}
	}
	panic("unknown face")
}
