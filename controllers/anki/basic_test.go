package anki

import (
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/model"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
)

func TestBasicButtons(t *testing.T) {
	tests := []struct {
		name     string
		card     *model.Card
		face     int
		expected studyview.ButtonMap
		err      string
	}{
		{
			name: "question",
			face: QuestionFace,
			expected: studyview.ButtonMap{
				"button-r": {Name: "Show Answer", Enabled: true},
			},
		},
		{
			name: "unsupported face",
			face: -1,
			err:  "Invalid face -1",
		},
		{
			name: "incorrect typed answer",
			card: &model.Card{
				Card: &fb.Card{
					Context: map[string]interface{}{
						contextKeyTypedAnswers: map[string]answer{
							"testField": {Text: "foo", Correct: false},
						},
					},
				},
			},
			face: AnswerFace,
			expected: studyview.ButtonMap{
				"button-l":  {Name: "Incorrect", Enabled: true},
				"button-cl": {Name: "Difficult", Enabled: false},
				"button-cr": {Name: "Correct", Enabled: false},
				"button-r":  {Name: "Easy", Enabled: false},
			},
		},
		{
			name: "normal answer",
			card: &model.Card{Card: &fb.Card{}},
			face: AnswerFace,
			expected: studyview.ButtonMap{
				"button-l":  {Name: "Incorrect", Enabled: true},
				"button-cl": {Name: "Difficult", Enabled: true},
				"button-cr": {Name: "Correct", Enabled: true},
				"button-r":  {Name: "Easy", Enabled: true},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := &Basic{}
			result, err := m.Buttons(test.card, test.face)
			testy.Error(t, test.err, err)
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestButtonsKey(t *testing.T) {
	tests := []struct {
		name     string
		card     *model.Card
		face     int
		expected string
	}{
		{
			name:     "question",
			face:     QuestionFace,
			expected: buttonsKeyQuestion,
		},
		{
			name:     "possibly correct answer",
			card:     &model.Card{Card: &fb.Card{}},
			face:     AnswerFace,
			expected: buttonsKeyAnswer,
		},
		{
			name: "incorrect answer",
			card: &model.Card{
				Card: &fb.Card{
					Context: map[string]interface{}{
						contextKeyTypedAnswers: map[string]answer{
							"testField": {Text: "foo", Correct: false},
						},
					},
				},
			},
			face:     AnswerFace,
			expected: buttonsKeyAnswerIncorrect,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := buttonsKey(test.card, test.face)
			if result != test.expected {
				t.Errorf("Unexpected result: %s", result)
			}
		})
	}
}
