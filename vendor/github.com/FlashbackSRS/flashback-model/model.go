package fb

import (
	"errors"
	"strconv"
)

// Model represents a Flashback card Model
type Model struct {
	Theme       *Theme `json:"-"`
	ID          uint32 `json:"id"` // I'd prefer uint8, but that doesn't support easy atomic incrementing
	Type        string `json:"modelType"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	// Templates are analogous to anki Card definitions. This field just holds the name of each template.
	Templates []string            `json:"templates"`
	Fields    []*Field            `json:"fields"`
	Files     *FileCollectionView `json:"files,omitempty"`
}

// Validate validates that all of the data in the theme appears valid and self
// consistent. A nil return value means no errors were detected.
func (m *Model) Validate() error {
	if m.Theme == nil {
		return errors.New("theme is required")
	}
	if m.Type == "" {
		return errors.New("type is required")
	}
	if m.Files == nil {
		return errors.New("file list must not be nil")
	}
	if m.Theme.Attachments == nil {
		return errors.New("invalid theme")
	}
	return nil
}

const (
	// AnkiStandardModel is a Basic Anki note
	AnkiStandardModel = "anki-basic"
	// AnkiClozeModel is an Anki Cloze note. Not yet implemented.
	AnkiClozeModel = "anki-cloze"
)

// NewModel creates a new model as a member of the provided theme, and of the
// provided type.
func NewModel(t *Theme, modelType string) (*Model, error) {
	if t == nil {
		return nil, errors.New("theme is required")
	}
	if modelType == "" {
		return nil, errors.New("model type is required")
	}
	return &Model{
		Theme:     t,
		ID:        t.NextModelSequence(),
		Type:      modelType,
		Templates: make([]string, 0, 1),
		Fields:    make([]*Field, 0, 1),
		Files:     t.Attachments.NewView(),
	}, nil
}

// AddFile adds a file of the provided name, type, and content as an attachment
// or returns an error.
func (m *Model) AddFile(name, ctype string, content []byte) error {
	return m.Files.AddFile(name, ctype, content)
}

// Identity returns the string representation of the model's identity.
func (m *Model) Identity() string {
	if m.Theme != nil {
		return m.Theme.Identity() + "." + strconv.FormatUint(uint64(m.ID), 16)
	}
	return ""
}

// AddField adds a field of the specified type and name to the Model.
func (m *Model) AddField(fType FieldType, name string) error {
	if fType > AnkiField {
		return errors.New("invalid field type")
	}
	if name == "" {
		return errors.New("field name is required")
	}
	m.Fields = append(m.Fields, &Field{
		Type: fType,
		Name: name,
	})
	return nil
}

// FieldType represents the valid field types
type FieldType int

const (
	// TextField is for a field which accepts only text
	TextField FieldType = iota
	// ImageField is for a field which accepts only an image
	ImageField
	// AudioField is for a field which accepts only audio
	AudioField
	// AnkiField is for a field improrted from Anki, which accepts Anki-specific
	// markup which may contain text, HTML, and other files.
	AnkiField
)

// Field represents a field of a model
type Field struct {
	Type FieldType `json:"fieldType"`
	Name string    `json:"name"` // Field name
}
