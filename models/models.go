package models

import "github.com/pkg/errors"

// A ModelHandler handles the logic for a given model
type ModelHandler interface {
	// Type returns the Model Type identifier string.
	Type() string
	// IframeScript returns a blob of JavaScript, which is loaded inside the
	// iframe of each card associated with this model type.
	IframeScript() []byte
	// Buttons returns the attributes for the three available answer buttons'
	// initial state. Index 0 = left button, 1 = center, 2 = right
	Buttons(face int) AnswerButtonsState
}

// AnswerButtonsState is the state of the three answer buttons
type AnswerButtonsState [3]AnswerButton

// AnswerButton is one of the three answer buttons.
type AnswerButton struct {
	Name    string
	Enabled bool
}

var handlers = map[string]ModelHandler{}
var handledTypes = []string{}

// RegisterModel registers a model handler
func RegisterModel(h ModelHandler) {
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
func GetHandler(modelType string) (ModelHandler, error) {
	h, ok := handlers[modelType]
	if !ok {
		return nil, errors.Errorf("no handler for '%s' registered", modelType)
	}
	return h, nil
}
