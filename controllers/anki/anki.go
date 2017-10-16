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

var (
	buttonsKeyQuestion        = fmt.Sprintf("%d", QuestionFace)
	buttonsKeyAnswer          = fmt.Sprintf("%d", AnswerFace)
	buttonsKeyAnswerIncorrect = fmt.Sprintf("%d-wrong", QuestionFace)
)

var buttonMaps = map[string]studyview.ButtonMap{
	buttonsKeyQuestion: studyview.ButtonMap{
		studyview.ButtonRight: {Name: "Show Answer", Enabled: true},
	},
	buttonsKeyAnswer: studyview.ButtonMap{
		studyview.ButtonLeft:        {Name: "Incorrect", Enabled: true},
		studyview.ButtonCenterLeft:  {Name: "Difficult", Enabled: true},
		studyview.ButtonCenterRight: {Name: "Correct", Enabled: true},
		studyview.ButtonRight:       {Name: "Easy", Enabled: true},
	},
	buttonsKeyAnswerIncorrect: studyview.ButtonMap{
		studyview.ButtonLeft:        {Name: "Incorrect", Enabled: true},
		studyview.ButtonCenterLeft:  {Name: "Difficult", Enabled: false},
		studyview.ButtonCenterRight: {Name: "Correct", Enabled: false},
		studyview.ButtonRight:       {Name: "Easy", Enabled: false},
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
