package note

import ()

type noteDoc struct {
	ID          string     `json:"_id"`
	Rev         string     `json:"_rev,omitempty"`
	Type        string     `json:"type"`
	Created     *time.Time `json:"created,omitempty"`
	Imported    *time.Time `json:"imported,omitempty"`
	Modified    *time.Time `json:"modified,omitempty"`
	ModelId     string     `json:"modelID"`
	Tags        []string   `json:"tags,omitempty"`
	FieldValues []string   `json:"fieldValues"`
	Comment     *string    `json:"comment,omitempty"`
	Attachments map[string]*model.Attachment `json:"_attachments"`
}

type Note struct {
	doc     noteDoc
	Comment string
}

func New() *Note {
}
