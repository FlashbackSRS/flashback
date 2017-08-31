package model

import (
	"testing"
	"time"

	"github.com/FlashbackSRS/flashback"
	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/flimzy/diff"
)

const floatTolerance = 0.00001

func float32sEqual(a, b float32) bool {
	return a-b < floatTolerance && b-a < floatTolerance
}

type ScheduleTest struct {
	Name             string
	Card             *fb.Card
	Now              time.Time
	Answer           flashback.AnswerQuality
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
			Now:              parseTime(t, "2017-01-01T00:00:00Z"),
			Answer:           flashback.AnswerPerfect,
			ExpectedEase:     2.5,
			ExpectedDue:      parseDue(t, "2017-01-02"),
			ExpectedInterval: flashback.InitialInterval,
		},
		ScheduleTest{
			Name:             "New card, Correct",
			Card:             &fb.Card{},
			Now:              parseTime(t, "2017-01-01T00:00:00Z"),
			Answer:           flashback.AnswerCorrect,
			ExpectedEase:     2.5,
			ExpectedDue:      parseDue(t, "2017-01-02"),
			ExpectedInterval: flashback.InitialInterval,
		},
		ScheduleTest{
			Name:             "New card, Difficult",
			Card:             &fb.Card{},
			Now:              parseTime(t, "2017-01-01T00:00:00Z"),
			Answer:           flashback.AnswerCorrectDifficult,
			ExpectedEase:     2.36,
			ExpectedDue:      parseDue(t, "2017-01-02"),
			ExpectedInterval: flashback.InitialInterval,
		},
		ScheduleTest{
			Name:             "New card, Wrong",
			Card:             &fb.Card{},
			Now:              parseTime(t, "2017-01-01T00:00:00Z"),
			Answer:           flashback.AnswerIncorrectEasy,
			ExpectedEase:     1.7,
			ExpectedDue:      parseDue(t, "2017-01-01 00:10:00"),
			ExpectedInterval: 10 * fb.Minute,
		},
		ScheduleTest{
			Name:             "New card, Wrong #2",
			Card:             &fb.Card{},
			Now:              parseTime(t, "2017-01-01T00:00:00Z"),
			Answer:           flashback.AnswerIncorrectRemembered,
			ExpectedEase:     1.7,
			ExpectedDue:      parseDue(t, "2017-01-01 00:10:00"),
			ExpectedInterval: 10 * fb.Minute,
		},

		// new card, failed once
		ScheduleTest{
			Name: "New card failed once, Easy",
			Card: &fb.Card{
				Due:        parseDue(t, "2017-01-02 00:10:00"),
				Interval:   10 * fb.Minute,
				EaseFactor: 1.7,
			},
			Now:              parseTime(t, "2017-01-01T00:10:00Z"),
			Answer:           flashback.AnswerPerfect,
			ExpectedEase:     1.8,
			ExpectedDue:      parseDue(t, "2017-01-02"),
			ExpectedInterval: flashback.InitialInterval,
		},
		ScheduleTest{
			Name: "New card failed once, Correct",
			Card: &fb.Card{
				Due:        parseDue(t, "2017-01-02"),
				Interval:   10 * fb.Minute,
				EaseFactor: 1.7,
			},
			Now:              parseTime(t, "2017-01-01T00:10:00Z"),
			Answer:           flashback.AnswerCorrect,
			ExpectedEase:     1.7,
			ExpectedDue:      parseDue(t, "2017-01-02"),
			ExpectedInterval: flashback.InitialInterval,
		},
		ScheduleTest{
			Name: "New card failed once, Difficult",
			Card: &fb.Card{
				Due:        parseDue(t, "2017-01-02 00:00:00"),
				Interval:   10 * fb.Minute,
				EaseFactor: 1.7,
			},
			Now:              parseTime(t, "2017-01-01T00:10:00Z"),
			Answer:           flashback.AnswerCorrectDifficult,
			ExpectedEase:     1.56,
			ExpectedDue:      parseDue(t, "2017-01-02"),
			ExpectedInterval: flashback.InitialInterval,
		},
		ScheduleTest{
			Name: "New card failed once, Wrong",
			Card: &fb.Card{
				Due:        parseDue(t, "2017-01-02 00:00:00"),
				Interval:   10 * fb.Minute,
				EaseFactor: 1.7,
			},
			Now:              parseTime(t, "2017-01-01T00:10:00Z"),
			Answer:           flashback.AnswerIncorrectEasy,
			ExpectedEase:     1.3,
			ExpectedDue:      parseDue(t, "2017-01-01 00:20:00"),
			ExpectedInterval: 10 * fb.Minute,
		},

		// Reviewed once
		ScheduleTest{
			Name: "Reviewed once, Easy",
			Card: &fb.Card{
				Due:         parseDue(t, "2017-01-02 00:00:00"),
				Interval:    flashback.InitialInterval,
				EaseFactor:  2.0,
				ReviewCount: 1,
			},
			Now:              parseTime(t, "2017-01-02T00:00:00Z"),
			Answer:           flashback.AnswerPerfect,
			ExpectedEase:     2.1,
			ExpectedDue:      parseDue(t, "2017-01-08"),
			ExpectedInterval: flashback.SecondInterval,
		},
		ScheduleTest{
			Name: "Reviewed once, Correct",
			Card: &fb.Card{
				Due:         parseDue(t, "2017-01-02 00:00:00"),
				Interval:    flashback.InitialInterval,
				EaseFactor:  2.0,
				ReviewCount: 1,
			},
			Now:              parseTime(t, "2017-01-02T00:00:00Z"),
			Answer:           flashback.AnswerCorrect,
			ExpectedEase:     2.0,
			ExpectedDue:      parseDue(t, "2017-01-08"),
			ExpectedInterval: flashback.SecondInterval,
		},
		ScheduleTest{
			Name: "Reviewed once, Difficult",
			Card: &fb.Card{
				Due:         parseDue(t, "2017-01-02 00:00:00"),
				Interval:    flashback.InitialInterval,
				EaseFactor:  2.0,
				ReviewCount: 1,
			},
			Now:              parseTime(t, "2017-01-02T00:00:00Z"),
			Answer:           flashback.AnswerCorrectDifficult,
			ExpectedEase:     1.86,
			ExpectedDue:      parseDue(t, "2017-01-08"),
			ExpectedInterval: flashback.SecondInterval,
		},
		ScheduleTest{
			Name: "Reviewed once, Wrong",
			Card: &fb.Card{
				Due:         parseDue(t, "2017-01-02 00:00:00"),
				Interval:    flashback.InitialInterval,
				EaseFactor:  2.0,
				ReviewCount: 1,
			},
			Now:              parseTime(t, "2017-01-02T00:00:00Z"),
			Answer:           flashback.AnswerIncorrectEasy,
			ExpectedEase:     1.3,
			ExpectedDue:      parseDue(t, "2017-01-02 00:10:00"),
			ExpectedInterval: 10 * fb.Minute,
		},
		// Reviewed once, 3 days late
		ScheduleTest{
			Name: "Reviewed once, 10 days late, Easy",
			Card: &fb.Card{
				Due:         parseDue(t, "2017-01-02"),
				Interval:    flashback.InitialInterval,
				EaseFactor:  2.0,
				ReviewCount: 1,
			},
			Now:              parseTime(t, "2017-01-04T00:00:00Z"),
			Answer:           flashback.AnswerPerfect,
			ExpectedEase:     2.1,
			ExpectedDue:      parseDue(t, "2017-01-11"),
			ExpectedInterval: 7 * fb.Day,
		},
		ScheduleTest{
			Name: "Real world #1",
			Now:  parseTime(t, "2017-01-23T12:56:21Z"),
			Card: &fb.Card{
				Due:         parseDue(t, "2017-01-24"),
				Interval:    fb.Day,
				EaseFactor:  2.36,
				ReviewCount: 0,
			},
			Answer:           flashback.AnswerCorrect,
			ExpectedEase:     2.36,
			ExpectedDue:      parseDue(t, "2017-01-24"),
			ExpectedInterval: fb.Day,
		},
	}
	for _, test := range tests {
		now = func() time.Time {
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
	Quality  flashback.AnswerQuality
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

func TestSchedule(t *testing.T) {
	tests := []struct {
		name     string
		card     *Card
		quality  flashback.AnswerQuality
		expected *Card
		err      string
	}{
		{
			name:    "new card, correct answer",
			card:    &Card{Card: &fb.Card{}},
			quality: flashback.AnswerCorrect,
			expected: &Card{Card: &fb.Card{
				LastReview:  now().UTC(),
				EaseFactor:  2.5,
				Interval:    86400000000000,
				Due:         fb.Due(now()).Add(86400000000000),
				BuriedUntil: fb.Due(now()).Add(86400000000000),
				ReviewCount: 1,
			}},
		},
		{
			name:    "new card, incorrect answer",
			card:    &Card{Card: &fb.Card{}},
			quality: flashback.AnswerBlackout,
			expected: &Card{Card: &fb.Card{
				EaseFactor:  1.7,
				Interval:    600000000000,
				Due:         fb.Due(now()).Add(600000000000),
				BuriedUntil: fb.Due(now()).Add(600000000000),
				ReviewCount: 0,
			}},
		},
		{
			name: "mature card, correct answer, bury for one day",
			card: &Card{Card: &fb.Card{
				EaseFactor:  2.5,
				Interval:    10 * fb.Day,
				ReviewCount: 5,
				Due:         fb.Due(now()),
			}},
			quality: flashback.AnswerCorrect,
			expected: &Card{Card: &fb.Card{
				LastReview:  now().UTC(),
				BuriedUntil: fb.Due(now().UTC()).Add(fb.Day),
				EaseFactor:  2.5,
				Interval:    2159999988006912,
				Due:         fb.Due(now()).Add(2159999988006912),
				ReviewCount: 6,
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := Schedule(test.card, time.Second, test.quality)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, test.card); d != nil {
				t.Error(d)
			}
		})
	}
}
