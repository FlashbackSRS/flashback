// Package mock is the model handler for testing purposes
package mock

// Model is an Anki Basic model
type Model struct{}

// Type returns the string "anki-basic", to identify this model handler's type.
func (m *Model) Type() string {
	return "mock-model"
}

// IframeScript returns JavaScript to run inside the iframe.
func (m *Model) IframeScript() []byte {
	return []byte(`
        /* Placeholder JS */
        console.log("Mock Handler");
    `)
}
