package cardmodel

import (
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback-model"
)

// A Model handles the logic for a given model
type Model interface {
	// Type returns the Model Type identifier string.
	Type() string
	// IframeScript returns a blob of JavaScript, which is loaded inside the
	// iframe of each card associated with this model type.
	IframeScript() []byte
	// Buttons returns the attributes for the three available answer buttons'
	// initial state. Index 0 = left button, 1 = center, 2 = right
	Buttons(face int) (AnswerButtonsState, error)
	// ButtonPress is called when the user presses a button. If done is returned
	// as true, the next card is selected. If done is false, the same card will
	// be displayed, with the current value of face (possibly changed by the
	// function)
	Action(card *fb.Card, face *int, action Action) (done bool, err error)
}

// An Action describes something that can happen after a button press
type Action struct {
	Button *Button
}

// Button represents a button displayed on each card face
type Button int

// The buttons displayed on each card
const (
	ButtonLeft Button = iota
	ButtonCenter
	ButtonRight
)

func (b Button) String() string {
	switch b {
	case ButtonLeft:
		return "Left"
	case ButtonCenter:
		return "Center"
	case ButtonRight:
		return "Right"
	}
	return "Unknown"
}

// AnswerButtonsState is the state of the three answer buttons
type AnswerButtonsState [3]AnswerButton

// AnswerButton is one of the three answer buttons.
type AnswerButton struct {
	Name    string
	Enabled bool
}

var handlers = map[string]Model{}
var handledTypes = []string{}

// RegisterModel registers a model handler
func RegisterModel(h Model) {
	name := h.Type()
	if _, ok := handlers[name]; ok {
		panic("A handler for '" + name + "' is already registered'")
	}
	handlers[name] = h
	handledTypes = append(handledTypes, name)
}

// HandledTypes returns a slice with the names of all handled types.
func HandledTypes() []string {
	return handledTypes
}

// GetHandler gets the handler for the specified model type
func GetHandler(modelType string) (Model, error) {
	h, ok := handlers[modelType]
	if !ok {
		return nil, errors.Errorf("no handler for '%s' registered", modelType)
	}
	return h, nil
}
