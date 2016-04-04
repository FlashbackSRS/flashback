package theme

import (
	"time"
)

type modelDoc struct {
	ID          string     `json:"_id"`
	Rev         string     `json:"_rev,omitempty"`
	Type        string     `json:"$Type"`
	Created     *time.Time `json:"$Created,omitempty"`
	Modified    *time.Time `json:"$Modified"`
	Imported    *time.Time `json:"$Imported,omitempty"`
	Name        *string    `json:"$Name"`
	Description *string    `json:"$Description,omitempty"`
}

type Model struct {
	doc         modelDoc
	theme       *Theme
	Name        string
	Description string
}

func NewModel(t *Theme) *Model {
	m := &Model{
		theme: t,
	}
	m.newModelDoc()
	return m
}

func (m *Model) newModelDoc() {
	m.doc = modelDoc{
		Name:        &m.Name,
		Description: &m.Description,
	}
}
