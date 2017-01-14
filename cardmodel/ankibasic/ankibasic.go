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
func (m *Model) Buttons(face int) (*cardmodel.ButtonMap, error) {
	switch face {
	case QuestionFace:
		return &cardmodel.ButtonMap{
			cardmodel.ButtonRight: cardmodel.AnswerButton{
				Name:    "Show Answer",
				Enabled: true,
			},
		}, nil
	case AnswerFace:
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
	default:
		return nil, errors.Errorf("Invalid face %d", face)
	}
}

// Action responds to a card action, such as a button press
func (m *Model) Action(card *fb.Card, face *int, action cardmodel.Action) (bool, error) {
	if action.Button == "" {
		return false, errors.New("Invalid response; no button press")
	}
	button := action.Button
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
