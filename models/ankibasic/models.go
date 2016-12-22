// Package ankibasic is the model handler for the Basic Anki model type.
package ankibasic

import "github.com/FlashbackSRS/flashback/models"

// Model is an Anki Basic model
type Model struct{}

func init() {
	models.RegisterModel(&Model{})
}

// Type returns the string "anki-basic", to identify this model handler's type.
func (m *Model) Type() string {
	return "anki-basic"
}

// IframeScript returns JavaScript to run inside the iframe.
func (m *Model) IframeScript() []byte {
	data, err := Asset("script.js")
	if err != nil {
		panic(err)
	}
	return data
}
