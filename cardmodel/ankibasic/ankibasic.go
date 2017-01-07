// Package ankibasic is the model handler for the Basic Anki model type.
package ankibasic

import (
	"github.com/flimzy/log"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback/cardmodel"
)

const (
	// FaceQuestion is the question face of a card
	FaceQuestion = iota
	// FaceAnswer is the answer face of a card
	FaceAnswer
)

// Model is an Anki Basic model
type Model struct{}

var _ cardmodel.Model = &Model{}

func init() {
	log.Debug("Registering anki-basic model\n")
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
func (m *Model) Buttons(face uint8) (cardmodel.AnswerButtonsState, error) {
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
		}, nil
	case FaceAnswer:
		return cardmodel.AnswerButtonsState{
			cardmodel.AnswerButton{
				Name:    "Wrong Answer",
				Enabled: true,
			},
			cardmodel.AnswerButton{
				Name:    "Mostly Correct",
				Enabled: true,
			},
			cardmodel.AnswerButton{
				Name:    "Correct Answer",
				Enabled: true,
			},
		}, nil
	default:
		return cardmodel.AnswerButtonsState{}, errors.Errorf("Invalid face %d", face)
	}
}
