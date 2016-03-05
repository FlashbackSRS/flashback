package data

import (
	"time"
)

const HTMLTemplateContentType = "text/html+flashbacktmpl"

type Attachment struct {
	Type string `json:"content_type"`
	MD5  string `json:"digest"`
	Stub bool   `json:"stub"`
}

type Model struct {
	Id           string                `json:"_id"`
	Rev          string                `json:"_rev,omitempty"`
	Attachments  map[string]Attachment `json:"_attachments,omitempty"`
	Type         string                `json:"$type"`
	Name         string                `json:"Name,omitempty"`
	Description  string                `json:"Description,omitempty"`
	Fields       []*Field              `json:"Fields"`
	Created      time.Time             `json:"$created,omitempty"`
	AnkiImported time.Time             `json:"$ankiImported,omitempty"` // Only for items imported from Anki
	Modified     time.Time             `json:"$modified,omitempty"`
	Comment      string                `json:"Comment,omitempty"`
}

type Field struct {
	Name string `json:"Name"`
}

type Deck struct {
	Id           string    `json:"_id"`
	Rev          string    `json:"_rev,omitempty"`
	Type         string    `json:"$type"`
	Name         string    `json:"Name,omitempty"`
	Description  string    `json:"Description,omitempty"`
	Created      time.Time `json:"$created,omitempty"`
	AnkiImported time.Time `json:"$ankiImported,omitempty"` // Only for items imported from Anki
	Modified     time.Time `json:"$modified,omitempty"`
	Comment      string    `json:"Comment"`
}

type DeckConfig struct {
	Id              string    `json:"_id"`
	Rev             string    `json:"_rev,omitempty"`
	Type            string    `json:"$type"`
	DeckId          string    `json:"$deckId"`
	Created         time.Time `json:"$created,omitempty"`
	AnkiImported    time.Time `json:"$ankiImported,omitempty"` // Only for items imported from Anki
	Modified        time.Time `json:"$modified,omitempty"`
	MaxDailyReviews int       `json:"MaxDailyReviews"`
	MaxDailyNew     int       `json:"MaxDailyNew"`
}

type Note struct {
	Id           string                `json:"_id"`
	Rev          string                `json:"_rev,omitempty"`
	Attachments  map[string]Attachment `json:"_attachments,omitempty"`
	Type         string                `json:"$type"`
	Created      time.Time             `json:"$created,omitempty"`
	AnkiImported time.Time             `json:"$ankiImported,omitempty"` // Only for items imported from Anki
	Modified     time.Time             `json:"$modified,omitempty"`
	ModelId      string                `json:"$modelId"`
	Tags         []string              `json:"Tags,omitempty"`
	FieldValues  []string              `json:"FieldValues"`
	Cards        []string              `json:"GeneratedCards"`
	Comment      string                `json:"Comment,omitempty"`
}

type Card struct {
	Id           string        `json:"_id"`
	Rev          string        `json:"_rev,omitempty"`
	Type         string        `json:"$type"`
	NoteId       string        `json:"$noteId"`
	DeckId       string        `json:"$deckId"`
	TemplateId   string        `json:"$templateId"`
	Created      *time.Time    `json:"$created,omitempty"`
	AnkiImported *time.Time    `json:"$ankiImported,omitempty"` // Only for items imported from Anki
	Modified     *time.Time    `json:"$modified,omitempty"`
	Due          *time.Time    `json:"Due,omitempty"`
	Reviews      int           `json:"Reviews"`
	Lapses       int           `json:"Lapses"`
	Interval     time.Duration `json:"Interval"`
	SRSFactor    float32       `json:"SRSFactor"`
	Suspended    bool          `json:"Suspended"`
	RelatedCards []string      `json:"Related"`
}

type Answer int

const (
	WrongAnswer Answer = iota
	HardAnswer
	OKAnswer
	EasyAnswer
)

type Review struct {
	Id           string        `json:"_id"`
	Rev          string        `json:"_rev,omitempty"`
	Type         string        `json:"$type"`
	CardId       string        `json:"$card"`
	Timestamp    time.Time     `json:"Timestamp"`
	Answer       string        `json:"Answer"`
	Interval     time.Duration `json:"Interval"`
	LastInterval time.Duration `json:"LastInterval"`
	Factor       float32       `json:"Factor"`
	ReviewType   string        `json:"Type"`
}
