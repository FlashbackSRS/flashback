package repo

import (
	"testing"
	"time"

	"github.com/FlashbackSRS/flashback-model"
)

const floatTolerance = 0.00001

func float32sEqual(a, b float32) bool {
	return a-b < floatTolerance && b-a < floatTolerance
}

func intervalPtr(d fb.Interval) *fb.Interval { return &d }
func parseDuePtr(src string) *fb.Due         { d := parseDue(src); return &d }

type ScheduleTest struct {
	Name             string
	Card             *fb.Card
	Now              time.Time
	Answer           AnswerQuality
	ExpectedEase     float32
	ExpectedDue      fb.Due
	ExpectedInterval fb.Interval
}

func TestScheduling(t *testing.T) {
	tests := []ScheduleTest{
		// Brand new card
		ScheduleTest{
			Name:             "New card, Easy",
			Card:             &fb.Card{},
			Now:              parseTime("2017-01-01 00:00:00"),
			Answer:           AnswerPerfect,
			ExpectedEase:     2.5,
			ExpectedDue:      parseDue("2017-01-02"),
			ExpectedInterval: InitialInterval,
		},
		ScheduleTest{
			Name:             "New card, Correct",
			Card:             &fb.Card{},
			Now:              parseTime("2017-01-01 00:00:00"),
			Answer:           AnswerCorrect,
			ExpectedEase:     2.5,
			ExpectedDue:      parseDue("2017-01-02"),
			ExpectedInterval: InitialInterval,
		},
		ScheduleTest{
			Name:             "New card, Difficult",
			Card:             &fb.Card{},
			Now:              parseTime("2017-01-01 00:00:00"),
			Answer:           AnswerCorrectDifficult,
			ExpectedEase:     2.36,
			ExpectedDue:      parseDue("2017-01-02"),
			ExpectedInterval: InitialInterval,
		},
		ScheduleTest{
			Name:             "New card, Wrong",
			Card:             &fb.Card{},
			Now:              parseTime("2017-01-01 00:00:00"),
			Answer:           AnswerIncorrectEasy,
			ExpectedEase:     1.7,
			ExpectedDue:      parseDue("2017-01-01 00:10:00"),
			ExpectedInterval: 10 * fb.Minute,
		},
		ScheduleTest{
			Name:             "New card, Wrong #2",
			Card:             &fb.Card{},
			Now:              parseTime("2017-01-01 00:00:00"),
			Answer:           AnswerIncorrectRemembered,
			ExpectedEase:     1.7,
			ExpectedDue:      parseDue("2017-01-01 00:10:00"),
			ExpectedInterval: 10 * fb.Minute,
		},

		// new card, failed once
		ScheduleTest{
			Name: "New card failed once, Easy",
			Card: &fb.Card{
				Due:        parseDuePtr("2017-01-02 00:10:00"),
				Interval:   intervalPtr(10 * fb.Minute),
				EaseFactor: 1.7,
			},
			Now:              parseTime("2017-01-01 00:10:00"),
			Answer:           AnswerPerfect,
			ExpectedEase:     1.8,
			ExpectedDue:      parseDue("2017-01-02"),
			ExpectedInterval: InitialInterval,
		},
		ScheduleTest{
			Name: "New card failed once, Correct",
			Card: &fb.Card{
				Due:        parseDuePtr("2017-01-02"),
				Interval:   intervalPtr(10 * fb.Minute),
				EaseFactor: 1.7,
			},
			Now:              parseTime("2017-01-01 00:10:00"),
			Answer:           AnswerCorrect,
			ExpectedEase:     1.7,
			ExpectedDue:      parseDue("2017-01-02"),
			ExpectedInterval: InitialInterval,
		},
		ScheduleTest{
			Name: "New card failed once, Difficult",
			Card: &fb.Card{
				Due:        parseDuePtr("2017-01-02 00:00:00"),
				Interval:   intervalPtr(10 * fb.Minute),
				EaseFactor: 1.7,
			},
			Now:              parseTime("2017-01-01 00:10:00"),
			Answer:           AnswerCorrectDifficult,
			ExpectedEase:     1.56,
			ExpectedDue:      parseDue("2017-01-02"),
			ExpectedInterval: InitialInterval,
		},
		ScheduleTest{
			Name: "New card failed once, Wrong",
			Card: &fb.Card{
				Due:        parseDuePtr("2017-01-02 00:00:00"),
				Interval:   intervalPtr(10 * fb.Minute),
				EaseFactor: 1.7,
			},
			Now:              parseTime("2017-01-01 00:10:00"),
			Answer:           AnswerIncorrectEasy,
			ExpectedEase:     1.3,
			ExpectedDue:      parseDue("2017-01-01 00:20:00"),
			ExpectedInterval: 10 * fb.Minute,
		},

		// Reviewed once
		ScheduleTest{
			Name: "Reviewed once, Easy",
			Card: &fb.Card{
				Due:         parseDuePtr("2017-01-02 00:00:00"),
				Interval:    intervalPtr(InitialInterval),
				EaseFactor:  2.0,
				ReviewCount: 1,
			},
			Now:              parseTime("2017-01-02 00:00:00"),
			Answer:           AnswerPerfect,
			ExpectedEase:     2.1,
			ExpectedDue:      parseDue("2017-01-08"),
			ExpectedInterval: SecondInterval,
		},
		ScheduleTest{
			Name: "Reviewed once, Correct",
			Card: &fb.Card{
				Due:         parseDuePtr("2017-01-02 00:00:00"),
				Interval:    intervalPtr(InitialInterval),
				EaseFactor:  2.0,
				ReviewCount: 1,
			},
			Now:              parseTime("2017-01-02 00:00:00"),
			Answer:           AnswerCorrect,
			ExpectedEase:     2.0,
			ExpectedDue:      parseDue("2017-01-08"),
			ExpectedInterval: SecondInterval,
		},
		ScheduleTest{
			Name: "Reviewed once, Difficult",
			Card: &fb.Card{
				Due:         parseDuePtr("2017-01-02 00:00:00"),
				Interval:    intervalPtr(InitialInterval),
				EaseFactor:  2.0,
				ReviewCount: 1,
			},
			Now:              parseTime("2017-01-02 00:00:00"),
			Answer:           AnswerCorrectDifficult,
			ExpectedEase:     1.86,
			ExpectedDue:      parseDue("2017-01-08"),
			ExpectedInterval: SecondInterval,
		},
		ScheduleTest{
			Name: "Reviewed once, Wrong",
			Card: &fb.Card{
				Due:         parseDuePtr("2017-01-02 00:00:00"),
				Interval:    intervalPtr(InitialInterval),
				EaseFactor:  2.0,
				ReviewCount: 1,
			},
			Now:              parseTime("2017-01-02 00:00:00"),
			Answer:           AnswerIncorrectEasy,
			ExpectedEase:     1.3,
			ExpectedDue:      parseDue("2017-01-02 00:10:00"),
			ExpectedInterval: 10 * fb.Minute,
		},
		// Reviewed once, 3 days late
		ScheduleTest{
			Name: "Reviewed once, 10 days late, Easy",
			Card: &fb.Card{
				Due:         parseDuePtr("2017-01-02"),
				Interval:    intervalPtr(InitialInterval),
				EaseFactor:  2.0,
				ReviewCount: 1,
			},
			Now:              parseTime("2017-01-04 00:00:00"),
			Answer:           AnswerPerfect,
			ExpectedEase:     2.1,
			ExpectedDue:      parseDue("2017-01-11"),
			ExpectedInterval: 7 * fb.Day,
		},
		ScheduleTest{
			Name: "Real world #1",
			Now:  parseTime("2017-01-23 12:56:21"),
			Card: &fb.Card{
				Due:         parseDuePtr("2017-01-24"),
				Interval:    intervalPtr(fb.Day),
				EaseFactor:  2.36,
				ReviewCount: 0,
			},
			Answer:           AnswerCorrect,
			ExpectedEase:     2.36,
			ExpectedDue:      parseDue("2017-01-24"),
			ExpectedInterval: fb.Day,
		},
	}
	for _, test := range tests {
		Now = func() time.Time {
			return test.Now
		}
		rcard := &Card{
			Card: test.Card,
		}
		ivl, ease := schedule(rcard, test.Answer)
		due := fb.Due(test.Now).Add(ivl)
		if !due.Equal(test.ExpectedDue) {
			t.Errorf("%s / Due:\n\tExpected: %s\n\t  Actual: %s\n", test.Name, test.ExpectedDue, due)
		}
		if !ivl.Equal(test.ExpectedInterval) {
			t.Errorf("%s / Interval:\n\tExpected: %s (%s)\n\t  Actual: %s (%s)\n", test.Name, test.ExpectedInterval, time.Duration(test.ExpectedInterval), ivl, time.Duration(ivl))
		}
		if !float32sEqual(ease, test.ExpectedEase) {
			t.Errorf("%s / Ease:\n\tExpected: %f\n\t  Actual: %f\n", test.Name, test.ExpectedEase, ease)
		}
	}
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
