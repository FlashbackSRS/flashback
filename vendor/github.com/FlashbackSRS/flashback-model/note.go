package fb

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/pkg/errors"
)

/*
type Note struct {
    Tags           string           `db:"tags"` // List of the note's tags
    UniqueField    string           `db:"sfld"` // The text of the first field, used for Anki's simplistic uniqueness checking
    Checksum       int64            `db:"csum"` // Field checksum used for duplicate check. Integer representation of first 8 digits of sha1 hash of the first field
}
*/

// Note represents a Flashback note.
type Note struct {
	ID          string          `json:"_id"`
	Rev         string          `json:"_rev,omitempty"`
	Created     time.Time       `json:"created"`
	Modified    time.Time       `json:"modified"`
	Imported    time.Time       `json:"imported,omitempty"`
	ThemeID     string          `json:"theme"`
	ModelID     uint32          `json:"model"`
	FieldValues []*FieldValue   `json:"fieldValues"`
	Attachments *FileCollection `json:"_attachments,omitempty"`
	Model       *Model          `json:"-"`
	// Set to true by UnmarshalJSON, to skip certain validation checks
	unmarshaling bool
}

// SetModel assigns the provided model to the Note. This is useful after retrieving
// a note.
func (n *Note) SetModel(m *Model) error {
	if err := n.validateModel(m); err != nil {
		return err
	}
	n.Model = m
	for i := 0; i < len(n.FieldValues); i++ {
		n.FieldValues[i].field = m.Fields[i]
	}
	return nil
}

func (n *Note) validateModel(m *Model) error {
	if m == nil {
		return errors.New("model required")
	}
	if m.Theme == nil {
		return errors.New("model theme required")
	}
	if m.Theme.ID != n.ThemeID {
		return errors.New("theme IDs must match")
	}
	if len(n.FieldValues) != len(m.Fields) {
		return errors.New("model.Fields and node.FieldValues lengths must match")
	}
	return nil
}

// Validate validates that all of the data in the note  appears valid and self
// consistent. A nil return value means no errors were detected.
func (n *Note) Validate() error {
	if n.ID == "" {
		return errors.New("id required")
	}
	if !strings.HasPrefix(n.ID, "note-") {
		return errors.New("incorrect doc type")
	}
	if !n.unmarshaling {
		if err := n.validateModel(n.Model); err != nil {
			return err
		}
	}
	if n.Created.IsZero() {
		return errors.New("created time required")
	}
	if n.Modified.IsZero() {
		return errors.New("modified time required")
	}
	if n.Attachments == nil {
		return errors.New("attachments collection must not be nil")
	}
	for i, fv := range n.FieldValues {
		if !n.unmarshaling {
			switch n.Model.Fields[i].Type {
			case TextField:
				if fv.files != nil {
					return errors.Errorf("text field %d must not have file list", i)
				}
			case AudioField:
				if fv.Text != "" {
					return errors.Errorf("audio field %d must not have text", i)
				}
			case ImageField:
				if fv.Text != "" {
					return errors.Errorf("image field %d must not have text", i)
				}
			}
		}
		if fv != nil && fv.files != nil && !n.Attachments.hasMemberView(fv.files) {
			return errors.Errorf("field %d file list must be member of attachments collection", i)
		}
	}
	return nil
}

// NewNote creates a new, empty note with the provided ID and Model.
func NewNote(id string, model *Model) (*Note, error) {
	if model == nil {
		return nil, errors.New("model required")
	}
	n := &Note{
		ID:          id,
		ThemeID:     model.Theme.ID,
		ModelID:     model.ID,
		Created:     now().UTC(),
		Modified:    now().UTC(),
		FieldValues: make([]*FieldValue, len(model.Fields)),
		Attachments: NewFileCollection(),
		Model:       model,
	}
	if err := n.Validate(); err != nil {
		return nil, err
	}
	return n, nil
}

type noteAlias Note

// MarshalJSON implements the json.Marshaler interface for the Note type.
func (n *Note) MarshalJSON() ([]byte, error) {
	if err := n.Validate(); err != nil {
		return nil, err
	}
	doc := struct {
		noteAlias
		Type     string     `json:"type"`
		Imported *time.Time `json:"imported,omitempty"`
	}{
		Type:      "note",
		noteAlias: noteAlias(*n),
	}
	if !n.Imported.IsZero() {
		doc.Imported = &n.Imported
	}
	return json.Marshal(doc)
}

// UnmarshalJSON implements the json.Unmarshaler interface for the Note type.
func (n *Note) UnmarshalJSON(data []byte) error {
	doc := &noteAlias{}
	if err := json.Unmarshal(data, doc); err != nil {
		return errors.Wrap(err, "failed to unmarshal Note")
	}
	*n = Note(*doc)
	if n.Attachments == nil {
		n.Attachments = NewFileCollection()
	}
	for _, fv := range n.FieldValues {
		if fv.files != nil {
			if err := n.Attachments.AddView(fv.files); err != nil {
				return err
			}
		}
	}
	n.unmarshaling = true
	if err := n.Validate(); err != nil {
		return err
	}
	n.unmarshaling = false
	return nil
}

// GetFieldValue returns the requested FieldValue by index.
func (n *Note) GetFieldValue(ord int) *FieldValue {
	fv := n.FieldValues[ord]
	if fv == nil {
		fv = &FieldValue{
			field: n.Model.Fields[ord],
		}
		n.FieldValues[ord] = fv
	}
	if fv.field == nil {
		panic("nil field? Did you set the note's model after load?")
	}
	if fv.field.Type != TextField {
		fv.files = n.Attachments.NewView()
	}
	return fv
}

// Type returns the FieldType of the FieldValue.
func (fv *FieldValue) Type() FieldType {
	if fv.field == nil {
		panic("nil field? Did you set the note's model after load?")
	}
	return fv.field.Type
}

// FieldValue stores the value of a given field.
type FieldValue struct {
	field *Field
	Text  string `json:"text,omitempty"`
	files *FileCollectionView
}

type fieldValueAlias FieldValue

type jsonFieldValue struct {
	fieldValueAlias
	Files *FileCollectionView `json:"files,omitempty"`
}

// MarshalJSON implements the json.Marshaler interface for the FieldValue type.
func (fv *FieldValue) MarshalJSON() ([]byte, error) {
	doc := jsonFieldValue{
		fieldValueAlias: fieldValueAlias(*fv),
		Files:           fv.files,
	}
	return json.Marshal(doc)
}

// UnmarshalJSON implements the json.Unmarshaler interface for the FieldValue type.
func (fv *FieldValue) UnmarshalJSON(data []byte) error {
	doc := &jsonFieldValue{}
	if err := json.Unmarshal(data, doc); err != nil {
		return errors.Wrap(err, "failed to unmarshal FieldValue")
	}
	*fv = FieldValue(doc.fieldValueAlias)
	fv.files = doc.Files
	return nil
}

// AddFile adds a file of the specified name, type, and content, as an attachment
// to be used by the FieldValue.
func (fv *FieldValue) AddFile(name, ctype string, content []byte) error {
	if fv.field == nil {
		panic("nil field? Did you set the note's model after load?")
	}
	if fv.field.Type == TextField {
		return errors.New("Text fields do not support attachments")
	}
	return fv.files.AddFile(name, ctype, content)
}

// SetRev sets the Note's _rev attribute.
func (n *Note) SetRev(rev string) { n.Rev = rev }

// DocID returns the Note's _id attribute.
func (n *Note) DocID() string { return n.ID }

// ImportedTime returns the time the Note was imported, or nil.
func (n *Note) ImportedTime() time.Time { return n.Imported }

// ModifiedTime returns the time the Note was last modified.
func (n *Note) ModifiedTime() time.Time { return n.Modified }

// MergeImport attempts to merge i into n, returning true if successful, or
// false if no merge was necessary.
func (n *Note) MergeImport(i interface{}) (bool, error) {
	existing := i.(*Note)
	if n.ID != existing.ID {
		return false, errors.New("IDs don't match")
	}
	if n.Imported.IsZero() || existing.Imported.IsZero() {
		return false, errors.New("not an import")
	}
	if !n.Created.Equal(existing.Created) {
		return false, errors.New("Created timestamps don't match")
	}
	n.Rev = existing.Rev
	if n.Modified.After(existing.Modified) {
		// The new version is newer than the existing one, so update
		return true, nil
	}
	// The new version is older, so we need to use the version we just read
	n.Modified = existing.Modified
	n.Imported = existing.Imported
	n.ModelID = existing.ModelID
	n.FieldValues = existing.FieldValues
	n.Attachments = existing.Attachments
	n.Model = existing.Model
	return false, nil
}
