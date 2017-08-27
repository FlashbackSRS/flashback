package anki

import (
	"time"

	"github.com/FlashbackSRS/flashback/diff"
	"github.com/FlashbackSRS/flashback/model"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
	"github.com/flimzy/log"
	"github.com/pkg/errors"
)

// Basic is the controller for the Anki Basic model
type Basic struct{}

var _ model.ModelController = &Basic{}

// Type returns the string "anki-basic", to identify this model handler's type.
func (m *Basic) Type() string {
	return "anki-basic"
}

// IframeScript returns JavaScript to run inside the iframe.
func (m *Basic) IframeScript() []byte {
	data, err := Asset("script.js")
	if err != nil {
		panic(err)
	}
	return data
}

// Buttons returns the initial button state
func (m *Basic) Buttons(face int) (studyview.ButtonMap, error) {
	buttons, ok := buttonMaps[face]
	if !ok {
		return nil, errors.Errorf("Invalid face %d", face)
	}
	return buttons, nil
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
			return true, nil
		}
		return false, nil
	case AnswerFace:
		log.Debugf("Old schedule: Due %s, Interval: %s, Ease: %f, ReviewCount: %d\n", card.Due, card.Interval, card.EaseFactor, card.ReviewCount)
		if err := model.Schedule(card, time.Now().Sub(startTime), quality(button)); err != nil {
			return false, err
		}
		log.Debugf("New schedule: Due %s, Interval: %s, Ease: %f, ReviewCount: %d\n", card.Due, card.Interval, card.EaseFactor, card.ReviewCount)
		card.Context = nil // Clear any saved answers
		return true, nil
	}
	log.Printf("Unexpected face/button combo: %d / %+v\n", *face, button)
	return false, nil
}
