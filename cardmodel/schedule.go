package cardmodel

import (
	"fmt"
	"time"

	"github.com/FlashbackSRS/flashback-model"
)

// Now is an alias for time.Now
var Now = time.Now

func now() fb.Due {
	return fb.Due(Now())
}

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

// Schedule is called by the model when the question is answered.
func Schedule(card *fb.Card, answerTime time.Duration, q AnswerQuality) {
	ivl, ease := schedule(card, q)
	due := now().Add(ivl)
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

func schedule(card *fb.Card, quality AnswerQuality) (interval fb.Interval, easeFactor float32) {
	ease := card.EaseFactor
	if ease == 0.0 {
		ease = InitialEase
	}

	if quality <= AnswerIncorrectEasy {
		quality = 0
		return LapseInterval, adjustEase(ease, quality)
	}

	if card.ReviewCount == 0 {
		return InitialInterval, adjustEase(ease, quality)
	}

	ease = adjustEase(ease, quality)
	interval = *card.Interval
	lastReviewed := card.Due.Add(-interval)
	observedInterval := fb.Interval(float32(now().Sub(lastReviewed)) * ease)
	if card.ReviewCount == 1 && observedInterval < SecondInterval {
		return SecondInterval, ease
	}
	fmt.Printf("Last reviewed on %s\n", lastReviewed)
	fmt.Printf("interval = %s, observed = %s, second = %s\n", interval, observedInterval, SecondInterval)
	if observedInterval > interval {
		interval = observedInterval
	}

	return interval, ease
}
