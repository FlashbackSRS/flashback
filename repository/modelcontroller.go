package repo

import (
	"html/template"
	"time"

	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
)

// ModelController is an interface for the per-type model controllers.
type ModelController interface {
	// Type returns the Model Type identifier string.
	Type() string
	// IframeScript returns a blob of JavaScript, which is loaded inside the
	// iframe of each card associated with this model type.
	IframeScript() []byte
	// Buttons returns the attributes for the three available answer buttons'
	// initial state. Index 0 = left button, 1 = center, 2 = right
	Buttons(face int) (studyview.ButtonMap, error)
	// Action is called when the card submits the 'mainform' form. `query` is the
	// deserialized content of the form submission, which should include at minimum
	// `submit` key. If done is returned as true, the next card is selected. If
	// done is false, the same card will be displayed, with the current value
	// of face (possibly changed by the function)
	Action(card *PouchCard, face *int, startTime time.Time, query *js.Object) (done bool, err error)
}

// FuncMapper is an optional interface that a ModelController may fulfill.
type FuncMapper interface {
	// FuncMap returns a FuncMap which can be used when parsing the
	// HTML templates for this note type.
	FuncMap() template.FuncMap
}

var modelControllers = map[string]ModelController{}
var modelControllerTypes = []string{}

// RegisterModelController registers a model controller for use in the app.
// The passed controller's Type() must return a unique value.
func RegisterModelController(c ModelController) {
	mType := c.Type()
	if _, ok := modelControllers[mType]; ok {
		panic("A controller for '" + mType + "' is already registered'")
	}
	modelControllers[mType] = c
	modelControllerTypes = append(modelControllerTypes, mType)
}

// RegisteredModelControllers returns a list of registered controller type
func RegisteredModelControllers() []string {
	return modelControllerTypes
}

func (m *Model) getController() (ModelController, error) {
	mType := m.Type
	c, ok := modelControllers[mType]
	if !ok {
		return nil, errors.Errorf("no handler for '%s' registered", mType)
	}
	return c, nil
}

func (c *PouchCard) getModelController() (ModelController, error) {
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

// Schedule implements the default scheduler.
func Schedule(card *PouchCard, answerDelay time.Duration, quality AnswerQuality) error {
	ivl, ease := schedule(card, quality)
	due := now().Add(ivl)
	card.Due = &due
	card.Interval = &ivl
	card.EaseFactor = ease
	if quality <= AnswerIncorrectEasy {
		card.ReviewCount = 0
	} else {
		now := time.Now()
		card.LastReview = &now
		card.ReviewCount++
	}
	return nil
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

func schedule(card *PouchCard, quality AnswerQuality) (interval fb.Interval, easeFactor float32) {
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
	log.Debugf("Last reviewed on %s\n", lastReviewed)
	log.Debugf("interval = %s, observed = %s, second = %s\n", interval, observedInterval, SecondInterval)
	if observedInterval > interval {
		interval = observedInterval
	}

	return interval, ease
}
