package model

import (
	"time"

	"github.com/FlashbackSRS/flashback"
	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/flimzy/log"
)

// Schedule implements the default scheduler.
func Schedule(card *Card, answerDelay time.Duration, quality flashback.AnswerQuality) error {
	ivl, ease := schedule(card, quality)
	card.Due = fb.Due(now()).Add(ivl)
	card.Interval = ivl
	card.EaseFactor = ease
	if quality <= flashback.AnswerIncorrectEasy {
		card.ReviewCount = 0
	} else {
		card.LastReview = now().UTC()
		card.ReviewCount++
	}
	return nil
}

func schedule(card *Card, quality flashback.AnswerQuality) (interval fb.Interval, easeFactor float32) {
	ease := card.EaseFactor
	if ease == 0.0 {
		ease = flashback.InitialEase
	}

	if quality <= flashback.AnswerIncorrectEasy {
		quality = 0
		return flashback.LapseInterval, adjustEase(ease, quality)
	}

	if card.ReviewCount == 0 {
		return flashback.InitialInterval, adjustEase(ease, quality)
	}

	ease = adjustEase(ease, quality)
	interval = card.Interval
	lastReviewed := time.Time(card.Due.Add(-interval))
	observedInterval := fb.Interval(float32(now().Sub(lastReviewed)) * ease)
	if card.ReviewCount == 1 && observedInterval < flashback.SecondInterval {
		return flashback.SecondInterval, ease
	}
	log.Debugf("Last reviewed on %s\n", lastReviewed)
	log.Debugf("interval = %s, observed = %s, second = %s\n", interval, observedInterval, flashback.SecondInterval)
	if observedInterval > interval {
		interval = observedInterval
	}

	return interval, ease
}

func adjustEase(ease float32, q flashback.AnswerQuality) float32 {
	quality := float32(q)
	newEase := ease + (0.1 - (5-quality)*(0.08+(5-quality)*0.02))
	if newEase < flashback.MinEase {
		return flashback.MinEase
	}
	if newEase > flashback.MaxEase {
		return flashback.MaxEase
	}
	return newEase
}
