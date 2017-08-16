package model

import (
	"encoding/json"

	fb "github.com/FlashbackSRS/flashback-model"
)

// fbNote is a wrapper around *fb.Note.
type fbNote struct {
	*fb.Note
}

type jsNote struct {
	ID string `json:"id"`
}

// MarshalJSON marshals a Note for the benefit of javascript context in HTML
// templates.
func (n *fbNote) MarshalJSON() ([]byte, error) {
	note := &jsNote{
		ID: n.ID,
	}
	return json.Marshal(note)
}
