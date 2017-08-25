// Package anki is the model handler for the Anki models.
package anki

import (
	"time"

	"github.com/flimzy/log"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback"
	"github.com/FlashbackSRS/flashback/diff"
	"github.com/FlashbackSRS/flashback/model"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
)

// The possible faces of an Anki card
const (
	QuestionFace = iota
	AnswerFace
)

// Basic is the controller for the Anki Basic model
type Basic struct{}

var _ model.ModelController = &Basic{}

func init() {
	log.Debug("Registering anki models\n")
	model.RegisterModelController(&Basic{})
	model.RegisterModelController(&Cloze{})
	log.Debug("Done registering anki models\n")
}

// Type returns the string "anki-basic", to identify this model handler's type.
func (m *Basic) Type() string {
	return "anki-basic"
}

//go:generate go-bindata -pkg anki -nocompress -prefix files -o data.go files

// IframeScript returns JavaScript to run inside the iframe.
func (m *Basic) IframeScript() []byte {
	data, err := Asset("script.js")
	if err != nil {
		panic(err)
	}
	return data
}

var buttonMaps = map[int]studyview.ButtonMap{
	QuestionFace: studyview.ButtonMap{
		studyview.ButtonRight: studyview.ButtonState{
			Name:    "Show Answer",
			Enabled: true,
		},
	},
	AnswerFace: studyview.ButtonMap{
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
	},
}

// Buttons returns the initial button state
func (m *Basic) Buttons(face int) (studyview.ButtonMap, error) {
	buttons, ok := buttonMaps[face]
	if !ok {
		return nil, errors.Errorf("Invalid face %d", face)
	}
	return buttons, nil
}

type answer struct {
	Text    string `json:"text"`
	Correct bool   `json:"correct"`
}

// Action responds to a card action, such as a button press
func (m *Basic) Action(card *model.Card, face *int, startTime time.Time, payload interface{}) (bool, error) {
	query := convertQuery(payload)
	log.Debugf("Submit recieved for face %d: %v\n", *face, query)
	button := studyview.Button(query.Submit)
	log.Debugf("Button %s pressed\n", button)
	switch *face {
	case QuestionFace:
		// Any input is fine; the only options are the right button, or 'ENTER' in a text field.
	case AnswerFace:
		if _, valid := buttonMaps[*face][button]; !valid {
			return false, errors.Errorf("Unexpected button press %s", button)
		}
	default:
		return false, errors.Errorf("Unexpected face %d", *face)
	}
	switch *face {
	case QuestionFace:
		*face++
		typedAnswers := query.TypedAnswers
		if len(typedAnswers) > 0 {
			results := make(map[string]answer)
			for _, fieldName := range card.Fields() {
				if typedAnswer, ok := typedAnswers[fieldName]; ok {
					fv := card.FieldValue(fieldName)
					if fv == nil {
						panic("No field value for field")
					}
					correct, d := diff.Diff(fv.Text, typedAnswer)
					results[fieldName] = answer{
						Text:    d,
						Correct: correct,
					}
				}
			}
			card.Context = map[string]interface{}{
				"typedAnswers": results,
			}
			// if err := card.Save(); err != nil {
			// 	return true, errors.Wrap(err, "save typedAnswers to card state")
			// }
		}
		return false, nil
	case AnswerFace:
		log.Debugf("Old schedule: Due %s, Interval: %s, Ease: %f, ReviewCount: %d\n", card.Due, card.Interval, card.EaseFactor, card.ReviewCount)
		if err := model.Schedule(card, time.Now().Sub(startTime), quality(button)); err != nil {
			return false, err
		}
		log.Debugf("New schedule: Due %s, Interval: %s, Ease: %f, ReviewCount: %d\n", card.Due, card.Interval, card.EaseFactor, card.ReviewCount)
		card.Context = nil // Clear any saved answers
		// if err := card.Save(); err != nil {
		// 	return true, errors.Wrap(err, "save card state")
		// }
		return true, nil
	}
	log.Printf("Unexpected face/button combo: %d / %+v\n", *face, button)
	return false, nil
}

func quality(button studyview.Button) flashback.AnswerQuality {
	switch button {
	case studyview.ButtonLeft:
		return flashback.AnswerBlackout
	case studyview.ButtonCenterLeft:
		return flashback.AnswerCorrectDifficult
	case studyview.ButtonCenterRight:
		return flashback.AnswerCorrect
	case studyview.ButtonRight:
		return flashback.AnswerPerfect
	}
	return flashback.AnswerBlackout
}
