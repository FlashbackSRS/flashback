// Package anki is the model handler for the Anki models.
package anki

import (
	"fmt"
	"html/template"
	"net/url"

	"github.com/flimzy/log"

	"github.com/FlashbackSRS/flashback"
	"github.com/FlashbackSRS/flashback/model"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
)

// The possible faces of an Anki card
const (
	QuestionFace = iota
	AnswerFace
)

func init() {
	log.Debug("Registering anki models\n")
	model.RegisterModelController(&Basic{})
	model.RegisterModelController(&Cloze{})
	log.Debug("Done registering anki models\n")
}

//go:generate go-bindata -pkg anki -nocompress -prefix files -o data.go files

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

type answer struct {
	Text    string `json:"text"`
	Correct bool   `json:"correct"`
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

var defaultFuncMap = map[string]interface{}{
	"image": image,
	"audio": audio,
}

func image(name string) template.HTML {
	return template.HTML(fmt.Sprintf(`<img src="%s">`, url.PathEscape(name)))
}

func audio(name, ctype string) template.HTML {
	return template.HTML(fmt.Sprintf(`<audio src="%s" type="%s"></audio>`, url.PathEscape(name), ctype))
}
