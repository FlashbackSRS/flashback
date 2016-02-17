package data

import (
	"time"
)

const HTMLTemplateContentType = "text/html+flashbacktmpl"

type Model struct {
	Id          string   `json:"_id"`
	Rev         string   `json:"_rev"`
	Type        string   `json:"Type"`
	Name        string   `json:"Name"`
	Description string   `json:"Description"`
	Fields      []*Field `json:"Fields"`
}

type Field struct {
	Name string `json:"Name"`
}

type File struct {
	Id          string // uuid
	Filename    string
	ContentType string
	Contents    string
}

type Template struct {
	Id          string // uuid
	Name        string
	Attachments []File // Main file is index.html?
}

type Card struct {
	Id           string // combination of Note.id and a counter
	NoteId       string // References Note.Id
	TemplateId   string // References Template.Id
	RelatedCards []int
	RelatedNotes []int
}

type Target struct {
	Id string
}

type CardStats struct {
	Id        string // References Card.Id
	Due       time.Time
	LastSeen  time.Time
	Suspended bool
	Notes     string
}

type Note struct {
	Id              string // uuid
	Fields          []Field
	LearnableFields []string
	// 	RelatedNotes     []int
	// 	RelatedCards     []int
	// 	NoteDependencies []int
	// 	CardDependencies []int
	Tags []string
}
