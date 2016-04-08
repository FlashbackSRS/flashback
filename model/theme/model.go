package theme

import (
	"time"
)

type modelDoc struct {
	ID          string     `json:"_id"`
	Rev         string     `json:"_rev,omitempty"`
	Type        string     `json:"type"`
	Created     *time.Time `json:"created,omitempty"`
	Modified    *time.Time `json:"modified"`
	Imported    *time.Time `json:"imported,omitempty"`
	Name        *string    `json:"name"`
	Description *string    `json:"description,omitempty"`
	Filenames   *[]string  `json:"filenames"`
}

type Model struct {
	doc         modelDoc
	theme       *Theme
	Name        string
	Description string
	Filenames   []string
}

func NewModel(t *Theme) *Model {
	m := &Model{
		theme:     t,
		Filenames: make([]string, 0),
	}
	m.newModelDoc()
	return m
}

func (m *Model) newModelDoc() {
	m.doc = modelDoc{
		Name:        &m.Name,
		Description: &m.Description,
		Filenames:   &m.Filenames,
	}
}
