package flashback

import (
	fb "github.com/FlashbackSRS/flashback-model"
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
