package repo

import (
	"time"

	"github.com/flimzy/log"

	"github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/controllers"
	"github.com/FlashbackSRS/flashback/model"
)

func (m *Model) getController() (controllers.ModelController, error) {
	return controllers.GetModelController(m.Type)
}

func (c *PouchCard) getModelController() (controllers.ModelController, error) {
	m, err := c.Model()
	if err != nil {
		return nil, err
	}
	return m.getController()
}

/*
FIXME: This scheduling stuff probably doesn't belong here. But where?
*/

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

// Now is an alias for time.Now
var Now = time.Now

func now() fb.Due {
	return fb.Due(Now())
}

// Schedule implements the default scheduler.
func Schedule(card *PouchCard, answerDelay time.Duration, quality model.AnswerQuality) error {
	ivl, ease := schedule(card, quality)
	due := now().Add(ivl)
	card.Due = &due
	card.Interval = &ivl
	card.EaseFactor = ease
	if quality <= model.AnswerIncorrectEasy {
		card.ReviewCount = 0
	} else {
		now := time.Now()
		card.LastReview = &now
		card.ReviewCount++
	}
	return nil
}

func adjustEase(ease float32, q model.AnswerQuality) float32 {
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

func schedule(card *PouchCard, quality model.AnswerQuality) (interval fb.Interval, easeFactor float32) {
	ease := card.EaseFactor
	if ease == 0.0 {
		ease = InitialEase
	}

	if quality <= model.AnswerIncorrectEasy {
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
	log.Debugf("Last reviewed on %s\n", lastReviewed)
	log.Debugf("interval = %s, observed = %s, second = %s\n", interval, observedInterval, SecondInterval)
	if observedInterval > interval {
		interval = observedInterval
	}

	return interval, ease
}
