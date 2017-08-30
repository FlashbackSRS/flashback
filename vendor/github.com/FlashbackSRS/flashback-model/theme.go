package fb

import (
	"encoding/json"
	"strings"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
)

// Theme contains related visual representation elements.
//
// A theme should generally group all card types (models) of a standard visual
// theme. A theme may contain files that are shared across all included modules,
// such as a CSS theme, common graphics files, or even JavaScript. Within a
// theme must exist one or more models, each of which represents a specific card
// type, and which may additionally have its own attachments.
type Theme struct {
	ID            string              `json:"_id"`
	Rev           string              `json:"_rev,omitempty"`
	Created       time.Time           `json:"created"`
	Modified      time.Time           `json:"modified"`
	Imported      time.Time           `json:"imported,omitempty"`
	Name          string              `json:"name,omitempty"`
	Description   string              `json:"description,omitempty"`
	Models        []*Model            `json:"models,omitempty"`
	Attachments   *FileCollection     `json:"_attachments,omitempty"`
	Files         *FileCollectionView `json:"files,omitempty"`
	ModelSequence uint32              `json:"modelSequence"`
}

// Validate validates that all of the data in the theme, including its models,
// appears valid and self consistent. A nil return value means no errors were
// detected.
func (t *Theme) Validate() error {
	if t.ID == "" {
		return errors.New("id required")
	}
	if err := validateDocID(t.ID); err != nil {
		return err
	}
	if !strings.HasPrefix(t.ID, "theme-") {
		return errors.New("incorrect doc type")
	}
	if t.Created.IsZero() {
		return errors.New("created time required")
	}
	if t.Modified.IsZero() {
		return errors.New("modified time required")
	}
	if t.Attachments == nil {
		return errors.New("attachments collection must not be nil")
	}
	if t.Files == nil {
		return errors.New("file list must not be nil")
	}
	if !t.Attachments.hasMemberView(t.Files) {
		return errors.New("file list must be a member of attachments collection")
	}
	for _, m := range t.Models {
		if t.ModelSequence <= m.ID {
			return errors.New("modelSequence must be larger than existing model IDs")
		}
		if !t.Attachments.hasMemberView(m.Files) {
			return errors.Errorf("model %d file list must be a member of attachments collection", m.ID)
		}
		if err := m.Validate(); err != nil {
			return errors.Wrap(err, "invalid model")
		}
	}
	return nil
}

// NewTheme returns a new, bare-bones theme, with the specified ID.
func NewTheme(id string) (*Theme, error) {
	t := &Theme{
		ID:       id,
		Created:  now().UTC(),
		Modified: now().UTC(),
	}
	t.Attachments = NewFileCollection()
	t.Files = t.Attachments.NewView()
	t.Models = make([]*Model, 0, 1)
	if err := t.Validate(); err != nil {
		return nil, err
	}
	return t, nil
}

// SetFile sets an attachment with the requested name, type, and content, as
// part of the Theme, overwriting any attachment with the same name, if it exists.
func (t *Theme) SetFile(name, ctype string, content []byte) {
	t.Files.SetFile(name, ctype, content)
}

type themeAlias Theme

// MarshalJSON implements the json.Marshaler interface for the Theme type.
func (t *Theme) MarshalJSON() ([]byte, error) {
	if err := t.Validate(); err != nil {
		return nil, err
	}
	doc := struct {
		themeAlias
		Type     string     `json:"type"`
		Imported *time.Time `json:"imported,omitempty"`
	}{
		Type:       "theme",
		themeAlias: themeAlias(*t),
	}
	if !t.Imported.IsZero() {
		doc.Imported = &t.Imported
	}
	return json.Marshal(doc)
}

// NewModel returns a new model of the requested type.
func (t *Theme) NewModel(mType string) (*Model, error) {
	m, err := NewModel(t, mType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create model")
	}
	t.Models = append(t.Models, m)
	return m, nil
}

// UnmarshalJSON implements the json.Unmarshaler interface for the Theme type.
func (t *Theme) UnmarshalJSON(data []byte) error {
	doc := &themeAlias{}
	if err := json.Unmarshal(data, doc); err != nil {
		return errors.Wrap(err, "failed to unmarshal Theme")
	}
	*t = Theme(*doc)

	if t.Attachments == nil {
		return errors.New("invalid theme: no attachments")
	}
	if t.Files == nil {
		return errors.New("invalid theme: no file list")
	}

	if err := t.Attachments.AddView(t.Files); err != nil {
		return err
	}
	for _, m := range t.Models {
		if err := t.Attachments.AddView(m.Files); err != nil {
			return err
		}
		m.Theme = t
	}
	return t.Validate()
}

// NextModelSequence returns the next available model sequence, while also
// updating the internal counter.
func (t *Theme) NextModelSequence() uint32 {
	id := t.ModelSequence
	atomic.AddUint32(&t.ModelSequence, 1)
	return id
}

// SetRev sets the _rev attribute of the Theme.
func (t *Theme) SetRev(rev string) { t.Rev = rev }

// DocID returns the theme's _id
func (t *Theme) DocID() string { return t.ID }

// ImportedTime returns the time the Theme was imported, or nil
func (t *Theme) ImportedTime() time.Time { return t.Imported }

// ModifiedTime returns the time the Theme was last modified
func (t *Theme) ModifiedTime() time.Time { return t.Modified }

// MergeImport attempts to merge i into t and returns true if a merge occurred,
// or false if no merge was necessary.
func (t *Theme) MergeImport(i interface{}) (bool, error) {
	existing := i.(*Theme)
	if t.ID != existing.ID {
		return false, errors.New("IDs don't match")
	}
	if t.Imported.IsZero() || existing.Imported.IsZero() {
		return false, errors.New("not an import")
	}
	if !t.Created.Equal(existing.Created) {
		return false, errors.New("Created timestamps don't match")
	}
	t.Rev = existing.Rev
	if t.Modified.After(existing.Modified) {
		// The new version is newer than the existing one, so update
		return true, nil
	}
	// The new version is older, so we need to use the version we just read
	t.Name = existing.Name
	t.Description = existing.Description
	t.Models = existing.Models
	t.Attachments = existing.Attachments
	t.Files = existing.Files
	t.ModelSequence = existing.ModelSequence
	t.Modified = existing.Modified
	t.Imported = existing.Imported
	return false, nil
}

// Identity returns the identifying tag for the Theme.
func (t *Theme) Identity() string {
	return strings.TrimPrefix(t.ID, "theme-")
}
