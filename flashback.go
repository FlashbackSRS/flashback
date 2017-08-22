package flashback

import (
	"context"
	"time"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
)

// AnswerQuality represents the SM-2 quality of the answer. See here:
// https://www.supermemo.com/english/ol/sm2.htm
type AnswerQuality int

// Answer qualities are borrowed from the SM2 algorithm.
const (
	// Complete Blackout
	AnswerBlackout AnswerQuality = iota
	// incorrect response; the correct one remembered
	AnswerIncorrectRemembered
	// incorrect response; where the correct one seemed easy to recall
	AnswerIncorrectEasy
	// correct response recalled with serious difficulty
	AnswerCorrectDifficult
	// correct response after a hesitation
	AnswerCorrect
	// perfect response
	AnswerPerfect
)

// Ease factor options
const (
	InitialEase float32 = 2.5
	MaxEase     float32 = 2.5
	MinEase     float32 = 1.3
)

// Interval options
const (
	InitialInterval = fb.Interval(24 * fb.Hour)
	SecondInterval  = fb.Interval(6 * fb.Day)
)

// Lapse options
const (
	LapseInterval = fb.Interval(10 * fb.Minute)
)

// Card represents a generic card-like object.
type Card interface {
	DocID() string
	Buttons(face int) (studyview.ButtonMap, error)
	Body(ctx context.Context, face int) (body string, err error)
	Action(face *int, startTime time.Time, query interface{}) (done bool, err error)
	BuryRelated(ctx context.Context) error
}
