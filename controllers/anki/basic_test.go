package anki

import (
	"testing"

	"github.com/FlashbackSRS/flashback/model"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
	"github.com/flimzy/diff"
	"github.com/flimzy/testy"
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
				"button-r": studyview.ButtonState{Name: "Show Answer", Enabled: true},
			},
		},
		{
			name: "unsupported face",
			face: -1,
			err:  "Invalid face -1",
		},
		{
			name: "answer",
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
