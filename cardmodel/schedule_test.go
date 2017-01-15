package cardmodel

import (
	"testing"
	"time"

	"github.com/FlashbackSRS/flashback-model"
)

const floatTolerance = 0.00001

func float32sEqual(a, b float32) bool {
	return a-b < floatTolerance && b-a < floatTolerance
}

func durationPtr(d time.Duration) *time.Duration { return &d }
func parseTimePtr(src string) *time.Time         { t := parseTime(src); return &t }
func parseTime(src string) time.Time {
	t, err := time.Parse(time.RFC3339, src)
	if err != nil {
		panic(err)
	}
	return t
}

type ScheduleTest struct {
	Name             string
	Card             *fb.Card
	Now              time.Time
	Answer           AnswerQuality
	ExpectedEase     float32
	ExpectedDue      time.Time
	ExpectedInterval time.Duration
}

func TestScheduling(t *testing.T) {
	tests := []ScheduleTest{
		// Brand new card
		ScheduleTest{
			Name:             "New card, Easy",
			Card:             &fb.Card{},
			Now:              parseTime("2017-01-01T00:00:00Z"),
			Answer:           AnswerPerfect,
			ExpectedEase:     2.5,
			ExpectedDue:      parseTime("2017-01-02T00:00:00Z"),
			ExpectedInterval: InitialInterval,
		},
		ScheduleTest{
			Name:             "New card, Correct",
			Card:             &fb.Card{},
			Now:              parseTime("2017-01-01T00:00:00Z"),
			Answer:           AnswerCorrect,
			ExpectedEase:     2.5,
			ExpectedDue:      parseTime("2017-01-02T00:00:00Z"),
			ExpectedInterval: InitialInterval,
		},
		ScheduleTest{
			Name:             "New card, Difficult",
			Card:             &fb.Card{},
			Now:              parseTime("2017-01-01T00:00:00Z"),
			Answer:           AnswerCorrectDifficult,
			ExpectedEase:     2.36,
			ExpectedDue:      parseTime("2017-01-02T00:00:00Z"),
			ExpectedInterval: InitialInterval,
		},
		ScheduleTest{
			Name:             "New card, Wrong",
			Card:             &fb.Card{},
			Now:              parseTime("2017-01-01T00:00:00Z"),
			Answer:           AnswerIncorrectEasy,
			ExpectedEase:     1.7,
			ExpectedDue:      parseTime("2017-01-01T00:10:00Z"),
			ExpectedInterval: 10 * time.Minute,
		},
		ScheduleTest{
			Name:             "New card, Wrong #2",
			Card:             &fb.Card{},
			Now:              parseTime("2017-01-01T00:00:00Z"),
			Answer:           AnswerIncorrectRemembered,
			ExpectedEase:     1.7,
			ExpectedDue:      parseTime("2017-01-01T00:10:00Z"),
			ExpectedInterval: 10 * time.Minute,
		},

		// new card, failed once
		ScheduleTest{
			Name: "New card failed once, Easy",
			Card: &fb.Card{
				Due:        parseTimePtr("2017-01-02T00:10:00Z"),
				Interval:   durationPtr(10 * time.Minute),
				EaseFactor: 1.7,
			},
			Now:              parseTime("2017-01-01T00:10:00Z"),
			Answer:           AnswerPerfect,
			ExpectedEase:     1.8,
			ExpectedDue:      parseTime("2017-01-02T00:10:00Z"),
			ExpectedInterval: InitialInterval,
		},
		ScheduleTest{
			Name: "New card failed once, Correct",
			Card: &fb.Card{
				Due:        parseTimePtr("2017-01-02T00:00:00Z"),
				Interval:   durationPtr(10 * time.Minute),
				EaseFactor: 1.7,
			},
			Now:              parseTime("2017-01-01T00:10:00Z"),
			Answer:           AnswerCorrect,
			ExpectedEase:     1.7,
			ExpectedDue:      parseTime("2017-01-02T00:10:00Z"),
			ExpectedInterval: InitialInterval,
		},
		ScheduleTest{
			Name: "New card failed once, Difficult",
			Card: &fb.Card{
				Due:        parseTimePtr("2017-01-02T00:00:00Z"),
				Interval:   durationPtr(10 * time.Minute),
				EaseFactor: 1.7,
			},
			Now:              parseTime("2017-01-01T00:10:00Z"),
			Answer:           AnswerCorrectDifficult,
			ExpectedEase:     1.56,
			ExpectedDue:      parseTime("2017-01-02T00:10:00Z"),
			ExpectedInterval: InitialInterval,
		},
		ScheduleTest{
			Name: "New card failed once, Wrong",
			Card: &fb.Card{
				Due:        parseTimePtr("2017-01-02T00:00:00Z"),
				Interval:   durationPtr(10 * time.Minute),
				EaseFactor: 1.7,
			},
			Now:              parseTime("2017-01-01T00:10:00Z"),
			Answer:           AnswerIncorrectEasy,
			ExpectedEase:     1.3,
			ExpectedDue:      parseTime("2017-01-01T00:20:00Z"),
			ExpectedInterval: 10 * time.Minute,
		},

		// Reviewed once
		ScheduleTest{
			Name: "Reviewed once, Easy",
			Card: &fb.Card{
				Due:         parseTimePtr("2017-01-02T00:00:00Z"),
				Interval:    durationPtr(InitialInterval),
				EaseFactor:  2.0,
				ReviewCount: 1,
			},
			Now:              parseTime("2017-01-02T00:00:00Z"),
			Answer:           AnswerPerfect,
			ExpectedEase:     2.1,
			ExpectedDue:      parseTime("2017-01-08T00:00:00Z"),
			ExpectedInterval: SecondInterval,
		},
		ScheduleTest{
			Name: "Reviewed once, Correct",
			Card: &fb.Card{
				Due:         parseTimePtr("2017-01-02T00:00:00Z"),
				Interval:    durationPtr(InitialInterval),
				EaseFactor:  2.0,
				ReviewCount: 1,
			},
			Now:              parseTime("2017-01-02T00:00:00Z"),
			Answer:           AnswerCorrect,
			ExpectedEase:     2.0,
			ExpectedDue:      parseTime("2017-01-08T00:00:00Z"),
			ExpectedInterval: SecondInterval,
		},
		ScheduleTest{
			Name: "Reviewed once, Difficult",
			Card: &fb.Card{
				Due:         parseTimePtr("2017-01-02T00:00:00Z"),
				Interval:    durationPtr(InitialInterval),
				EaseFactor:  2.0,
				ReviewCount: 1,
			},
			Now:              parseTime("2017-01-02T00:00:00Z"),
			Answer:           AnswerCorrectDifficult,
			ExpectedEase:     1.86,
			ExpectedDue:      parseTime("2017-01-08T00:00:00Z"),
			ExpectedInterval: SecondInterval,
		},
		ScheduleTest{
			Name: "Reviewed once, Wrong",
			Card: &fb.Card{
				Due:         parseTimePtr("2017-01-02T00:00:00Z"),
				Interval:    durationPtr(InitialInterval),
				EaseFactor:  2.0,
				ReviewCount: 1,
			},
			Now:              parseTime("2017-01-02T00:00:00Z"),
			Answer:           AnswerIncorrectEasy,
			ExpectedEase:     1.3,
			ExpectedDue:      parseTime("2017-01-02T00:10:00Z"),
			ExpectedInterval: 10 * time.Minute,
		},
		// Reviewed once, 3 days late
		ScheduleTest{
			Name: "Reviewed once, 10 days late, Easy",
			Card: &fb.Card{
				Due:         parseTimePtr("2017-01-02T00:00:00Z"),
				Interval:    durationPtr(InitialInterval),
				EaseFactor:  2.0,
				ReviewCount: 1,
			},
			Now:              parseTime("2017-01-04T00:00:00Z"),
			Answer:           AnswerPerfect,
			ExpectedEase:     2.1,
			ExpectedDue:      parseTime("2017-01-10T07:11:59.962349568Z"),
			ExpectedInterval: SecondInterval,
		},
	}
	for _, test := range tests {
		Now = func() time.Time {
			return test.Now
		}
		due, ivl, ease := schedule(test.Card, test.Answer)
		if !due.Equal(test.ExpectedDue) {
			t.Errorf("%s / Due:\n\tExpected: %s\n\t  Actual: %s\n", test.Name, test.ExpectedDue, due)
		}
		if ivl != test.ExpectedInterval {
			t.Errorf("%s / Interval:\n\tExpectd: %s\n\t  Actual: %s\n", test.Name, test.ExpectedInterval, ivl)
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
