package data

import (
	"time"
)

const HTMLTemplateContentType = "text/html+flashbacktmpl"

type Model struct {
	Id          string    `json:"_id"`
	Rev         string    `json:"_rev"`
	Type        string    `json:"$type"`
	Name        string    `json:"Name,omitempty"`
	Description string    `json:"Description,omitempty"`
	Fields      []*Field  `json:"Fields"`
	Created     time.Time `json:"$created,omitempty"`
	Modified    time.Time `json:"$modified,omitempty"`
	Comment     string    `json:"Comment,omitempty"`
}

type Field struct {
	Name string `json:"Name"`
}

type Deck struct {
	Id          string    `json:"_id"`
	Rev         string    `json:"_rev"`
	Type        string    `json:"$type"`
	AnkiId      string    `json:"$ankiId"`
	Name        string    `json:"Name,omitempty"`
	Description string    `json:"Description,omitempty"`
	Created     time.Time `json:"$created,omitempty"`
	Modified    time.Time `json:"$modified,omitempty"`
	Comment     string    `json:"Comment"`
}

type DeckConfig struct {
	Id              string    `json:"_id"`
	Rev             string    `json:"_rev"`
	Type            string    `json:"$type"`
	DeckId          string    `json:"$deckId"`
	Created         time.Time `json:"$created,omitempty"`
	Modified        time.Time `json:"$modified,omitempty"`
	MaxDailyReviews uint16    `json:"MaxDailyReviews"`
	MaxDailyNew     uint16    `json:"MaxDailyNew"`
}

type Note struct {
	Id          string    `json:"_id"`
	Rev         string    `json:"_rev"`
	AnkiId      string    `json:"$ankiId"`
	Type        string    `json:"$type"`
	Created     time.Time `json:"$created,omitempty"`
	Modified    time.Time `json:"$modified,omitempty"`
	ModelId     string    `json:"$modelId"`
	Tags        []string  `json:"Tags"`
	FieldValues []string  `json:"FieldValues"`
}

type Card struct {
	Id           string    `json:"_id"`
	Rev          string    `json:"_rev"`
	AnkiId       string    `json:"$ankiId"`
	Type         string    `json:"$type"`
	NoteId       string    `json:"$noteId"`
	DeckId       string    `json:"$deckId"`
	TemplateId   string    `json:"$templateId"`
	Created      time.Time `json:"$created,omitempty"`
	Modified     time.Time `json:"$modified,omitempty"`
	Due          time.Time `json:"Due,omitempty"`
	Reviews      int       `json:"Reviews"`
	Lapses       int       `json:"Lapses"`
	Interval     int       `json:"Interval"`
	SRSFactor    float32   `json:"SRSFactor"`
	Suspended    bool      `json:"Suspended"`
	RelatedCards []string  `json:"Related"`
}
