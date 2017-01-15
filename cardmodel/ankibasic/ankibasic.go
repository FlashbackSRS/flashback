// Package ankibasic is the model handler for the Basic Anki model type.
package ankibasic

import (
	"time"

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

var buttonMaps = map[int]cardmodel.ButtonMap{
	QuestionFace: cardmodel.ButtonMap{
		cardmodel.ButtonRight: cardmodel.AnswerButton{
			Name:    "Show Answer",
			Enabled: true,
		},
	},
	AnswerFace: cardmodel.ButtonMap{
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
	},
}

// Buttons returns the initial button state
func (m *Model) Buttons(face int) (cardmodel.ButtonMap, error) {
	buttons, ok := buttonMaps[face]
	if !ok {
		return nil, errors.Errorf("Invalid face %d", face)
	}
	return buttons, nil
}

// Action responds to a card action, such as a button press
func (m *Model) Action(card *fb.Card, face *int, startTime time.Time, action cardmodel.Action) (bool, error) {
	if action.Button == "" {
		return false, errors.New("Invalid response; no button press")
	}
	button := action.Button
	log.Debugf("%s button pressed for face %d\n", button, *face)
	if btns, ok := buttonMaps[*face]; ok {
		if _, valid := btns[button]; !valid {
			return false, errors.Errorf("Unexpected button press %s", button)
		}
	} else {
		return false, errors.Errorf("Unexpected face %d", *face)
	}
	switch *face {
	case QuestionFace:
		*face++
		return false, nil
	case AnswerFace:
		log.Debugf("Old schedule: Due %s, Interval: %s, Ease: %f\n", card.Due, card.Interval, card.EaseFactor)
		cardmodel.Schedule(card, time.Now().Sub(startTime), quality(button))
		log.Debugf("New schedule: Due %s, Interval: %s, Ease: %f\n", card.Due, card.Interval, card.EaseFactor)
		return true, nil
	}
	log.Printf("Unexpected face/action combo: %d / %+v\n", *face, action)
	return false, nil
}

func quality(button cardmodel.Button) cardmodel.AnswerQuality {
	switch button {
	case cardmodel.ButtonLeft:
		return cardmodel.AnswerBlackout
	case cardmodel.ButtonCenterLeft:
		return cardmodel.AnswerCorrectDifficult
	case cardmodel.ButtonCenterRight:
		return cardmodel.AnswerCorrect
	case cardmodel.ButtonRight:
		return cardmodel.AnswerPerfect
	}
	return cardmodel.AnswerBlackout
}
