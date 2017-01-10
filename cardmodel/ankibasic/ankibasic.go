// Package ankibasic is the model handler for the Basic Anki model type.
package ankibasic

import (
	"github.com/flimzy/log"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/cardmodel"
)

// The possible faces of an Anki card
const (
	QuestionFace = iota
	AnswerFace
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
func (m *Model) Buttons(face int) (cardmodel.AnswerButtonsState, error) {
	switch face {
	case QuestionFace:
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
	case AnswerFace:
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

// Action responds to a card action, such as a button press
func (m *Model) Action(card *fb.Card, face *int, action cardmodel.Action) (bool, error) {
	if action.Button == nil {
		return false, errors.New("Invalid response; no button press")
	}
	button := *action.Button
	log.Debugf("%s button pressed for face %d\n", button, face)
	switch *face {
	case QuestionFace:
		if button != cardmodel.ButtonRight {
			return false, errors.Errorf("Unexpected button press %s", button)
		}
		*face++
		return false, nil
	case AnswerFace:
		if button < cardmodel.ButtonLeft || button > cardmodel.ButtonRight {
			return false, errors.Errorf("Unexpected button press %s", button)
		}
		return true, nil
	}
	log.Printf("Unexpected face/action combo: %d / %+v\n", *face, action)
	return false, nil
}
