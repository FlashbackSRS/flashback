package cardmodel

import (
	"time"

	"github.com/FlashbackSRS/flashback-model"
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

const (
	InitialEase float32 = 2.5
	MaxEase     float32 = 2.5
	MinEase     float32 = 1.3
)

const (
	InitialInterval = 24 * time.Hour
	SecondInterval  = 6 * 24 * time.Hour
)

const LapseInterval = 10 * time.Minute

// Schedule is called by the model when the question is answered.
func Schedule(card *fb.Card, q AnswerQuality) {
	due, ivl, ease := schedule(card, q)
	card.Due = &due
	card.Interval = &ivl
	card.EaseFactor = ease
	if q <= AnswerIncorrectEasy {
		card.ReviewCount = 0
	}
}

func adjustEase(ease float32, q AnswerQuality) float32 {
	quality := float32(q)
	newEase := ease + (0.1 - (5-quality)*(0.08+(5-quality)*0.02))
	if newEase < MinEase {
		return MinEase
	}
	if newEase > MaxEase {
		return MaxEase
	}
	return newEase
}

func schedule(card *fb.Card, quality AnswerQuality) (due time.Time, interval time.Duration, easeFactor float32) {
	ease := card.EaseFactor
	if ease == 0.0 {
		ease = InitialEase
	}

	if quality <= AnswerIncorrectEasy {
		return time.Now().Add(LapseInterval), LapseInterval, adjustEase(ease, quality)
	}

	if card.ReviewCount == 0 {
		return time.Now().Add(InitialInterval), InitialInterval, adjustEase(ease, quality)
	}

	// if card.ReviewCount == 1 {
	// 	ivl := card.Due.Sub(*card.Interval)
	// 	if ivl.After(time.Now()) {
	//
	// 	}
	// 	due := time.Now().Add(SecondInterval)
	//
	// }
	return
}
