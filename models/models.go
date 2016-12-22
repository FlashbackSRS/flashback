package models

import "errors"

// A ModelHandler handles the logic for a given model
type ModelHandler interface {
	// Type returns the Model Type identifier string.
	Type() string
	// IframeScript returns a blob of JavaScript, which is loaded inside the
	// iframe of each card associated with this model type.
	IframeScript() []byte
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
		return nil, errors.New("no such handler registered")
	}
	return h, nil
}
