package cardmodel

import (
	"testing"

	"github.com/FlashbackSRS/flashback-model"
)

type ScheduleTest struct {
	Name         string
	Card         *fb.Card
	Answer       AnswerQuality
	ExpectedEase float32
}

func xTestScheduling(t *testing.T) {

}

type EaseTest struct {
	Ease     float32
	Quality  AnswerQuality
	Expected float32
}

func TestAdjustEase(t *testing.T) {
	tests := []EaseTest{
		EaseTest{
			Ease:     2.0,
			Quality:  5,
			Expected: 2.1,
		},
		EaseTest{
			Ease:     2.0,
			Quality:  4,
			Expected: 2.0,
		},
		EaseTest{
			Ease:     2.0,
			Quality:  3,
			Expected: 1.86,
		},
		EaseTest{
			Ease:     2.0,
			Quality:  2,
			Expected: 1.68,
		},
		EaseTest{
			Ease:     2.0,
			Quality:  1,
			Expected: 1.46,
		},
		EaseTest{
			Ease:     2.0,
			Quality:  0,
			Expected: 1.3,
		},
	}
	for _, test := range tests {
		result := adjustEase(test.Ease, test.Quality)
		if result != test.Expected {
			t.Errorf("Unexpected Ease result for %f/%d\n\tExpected: %f\n\t  Actual: %f\n", test.Ease, test.Quality, test.Expected, result)
		}
	}
}
