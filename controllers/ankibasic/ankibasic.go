// Package ankibasic is the model handler for the Basic Anki model type.
package ankibasic

import (
	"fmt"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback/diff"
	repo "github.com/FlashbackSRS/flashback/repository"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
)

// The possible faces of an Anki card
const (
	QuestionFace = iota
	AnswerFace
)

// AnkiBasic is the controller for the Anki Basic model
type AnkiBasic struct{}

var _ repo.ModelController = &AnkiBasic{}

func init() {
	log.Debug("Registering anki-basic model\n")
	repo.RegisterModelController(&AnkiBasic{})
}

// Type returns the string "anki-basic", to identify this model handler's type.
func (m *AnkiBasic) Type() string {
	return "anki-basic"
}

// IframeScript returns JavaScript to run inside the iframe.
func (m *AnkiBasic) IframeScript() []byte {
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
func (m *AnkiBasic) Buttons(face int) (studyview.ButtonMap, error) {
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
func (m *AnkiBasic) Action(card *repo.PouchCard, face *int, startTime time.Time, query *js.Object) (bool, error) {
	log.Debugf("Submit recieved for face %d: %v\n", *face, query)
	button := studyview.Button(query.Get("submit").String())
	log.Debugf("Button %s pressed\n", button)
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
		typedAnswers := make(map[string]string)
		for _, k := range js.Keys(query) {
			if strings.HasPrefix(k, "type:") {
				typedAnswers[k] = query.Get(k).String()
			}
		}
		if len(typedAnswers) > 0 {
			results := make(map[string]answer)
			m, err := card.Model()
			if err != nil {
				return false, errors.Wrap(err, "failed to get model")
			}
			n, err := card.Note()
			if err != nil {
				return false, errors.Wrap(err, "failed to get note")
			}
			for i, field := range m.Fields {
				fmt.Printf("field.Name = %s\n", field.Name)
				if typedAnswer, ok := typedAnswers["type:"+field.Name]; ok {
					fmt.Printf("Found one\n")
					correctAnswer, err := n.FieldValues[i].Text()
					if err != nil {
						panic("no text for typed answer !!")
					}
					correct, d := diff.Diff(correctAnswer, typedAnswer)
					results[field.Name] = answer{
						Text:    d,
						Correct: correct,
					}
				}
			}
			spew.Dump(results)
			card.Context = map[string]interface{}{
				"typedAnswers": results,
			}
			if err := card.Save(); err != nil {
				return true, errors.Wrap(err, "save typedAnswers to card state")
			}
		}
		return false, nil
	case AnswerFace:
		log.Debugf("Old schedule: Due %s, Interval: %s, Ease: %f, ReviewCount: %d\n", card.Due, card.Interval, card.EaseFactor, card.ReviewCount)
		repo.Schedule(card, time.Now().Sub(startTime), quality(button))
		log.Debugf("New schedule: Due %s, Interval: %s, Ease: %f, ReviewCount: %d\n", card.Due, card.Interval, card.EaseFactor, card.ReviewCount)
		card.Context = nil // Clear any saved answers
		if err := card.Save(); err != nil {
			return true, errors.Wrap(err, "save card state")
		}
		return true, nil
	}
	log.Printf("Unexpected face/button combo: %d / %+v\n", *face, button)
	return false, nil
}

func quality(button studyview.Button) repo.AnswerQuality {
	switch button {
	case studyview.ButtonLeft:
		return repo.AnswerBlackout
	case studyview.ButtonCenterLeft:
		return repo.AnswerCorrectDifficult
	case studyview.ButtonCenterRight:
		return repo.AnswerCorrect
	case studyview.ButtonRight:
		return repo.AnswerPerfect
	}
	return repo.AnswerBlackout
}
