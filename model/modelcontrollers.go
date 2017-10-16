package model

import (
	"errors"
	"html/template"
	"time"

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
	Buttons(card *Card, face int) (studyview.ButtonMap, error)
	// Action is called when the card submits the 'mainform' form. `query` is the
	// deserialized content of the form submission, which should include at minimum
	// `submit` key. If done is returned as true, the next card is selected. If
	// done is false, the same card will be displayed, with the current value
	// of face (possibly changed by the function)
	Action(card *Card, face *int, startTime time.Time, query interface{}) (done bool, err error)
}

// FuncMapper is an optional interface that a ModelController may fulfill.
type FuncMapper interface {
	// FuncMap is given the card and face, and is expected to return a
	// template.FuncMap which can be used when parsing the HTML templates for
	// this note type. If card is nil, the function is fine to return only
	// non-functional, placeholder methods.
	FuncMap(card *Card, face int) template.FuncMap
}

var modelControllers = map[string]ModelController{}
var modelControllerTypes = []string{}

// RegisterModelController registers a model controller for use in the app.
// The passed controller's Type() must return a unique value.
func RegisterModelController(c ModelController) {
	mType := c.Type()
	if _, ok := modelControllers[mType]; ok {
		panic("A controller for '" + mType + "' is already registered")
	}
	modelControllers[mType] = c
	modelControllerTypes = append(modelControllerTypes, mType)
}

// RegisteredModelControllers returns a list of registered controller type
func RegisteredModelControllers() []string {
	return modelControllerTypes
}

// GetModelController returns the ModelController for the registered model type,
// 'mType'.
func GetModelController(mType string) (ModelController, error) {
	if c, ok := modelControllers[mType]; ok {
		return c, nil
	}
	return nil, errors.New("ModelController for '" + mType + "' not found")
}
