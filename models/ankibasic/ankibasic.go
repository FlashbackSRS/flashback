// Package ankibasic is the model handler for the Basic Anki model type.
package ankibasic

import "github.com/FlashbackSRS/flashback/models"

const (
	// FaceQuestion is the question face of a card
	FaceQuestion = iota
	// FaceAnswer is the answer face of a card
	FaceAnswer
)

// Model is an Anki Basic model
type Model struct{}

func init() {
	models.RegisterModel(&Model{})
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
func (m *Model) Buttons(face int) models.AnswerButtonsState {
	switch face {
	case FaceQuestion:
		return models.AnswerButtonsState{
			models.AnswerButton{
				Name:    "",
				Enabled: false,
			},
			models.AnswerButton{
				Name:    "",
				Enabled: false,
			},
			models.AnswerButton{
				Name:    "Show Answer",
				Enabled: true,
			},
		}
	case FaceAnswer:
		return models.AnswerButtonsState{
			models.AnswerButton{
				Name:    "Wrong Answer",
				Enabled: false,
			},
			models.AnswerButton{
				Name:    "Mostly Correct",
				Enabled: false,
			},
			models.AnswerButton{
				Name:    "Correct Answer",
				Enabled: true,
			},
		}
	}
	panic("unknown face")
}
