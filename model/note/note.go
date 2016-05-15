package note

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/flimzy/go-pouchdb"

	"github.com/flimzy/flashback/anki"
	"github.com/flimzy/flashback/model"
	"github.com/flimzy/flashback/model/theme"
	"github.com/flimzy/flashback/model/user"
)

type noteDoc struct {
	ID          string                       `json:"_id"`
	Rev         string                       `json:"_rev,omitempty"`
	Type        string                       `json:"type"`
	Created     *time.Time                   `json:"created,omitempty"`
	Imported    *time.Time                   `json:"imported,omitempty"`
	Modified    *time.Time                   `json:"modified,omitempty"`
	ModelID     string                       `json:"modelID"`
	Tags        []string                     `json:"tags,omitempty"`
	FieldValues []string                     `json:"fieldValues"`
	Comment     *string                      `json:"comment,omitempty"`
	Attachments map[string]*model.Attachment `json:"_attachments"`
}

type Note struct {
	doc     noteDoc
	Theme   *theme.Theme
	Comment string
}

func New(t *theme.Theme, ModelID int) *Note {
	n := Note{
		Theme: t,
	}
	n.newNoteDoc()
	now := time.Now()
	n.doc.Modified = &now
	n.doc.Created = &now
	n.doc.ModelID = n.baseModelID() + "/" + strconv.Itoa(ModelID)
	return &n
}

func (n *Note) baseModelID() string {
	return strings.TrimPrefix(n.Theme.ID(), "theme-")
}

func (n *Note) setID(subtype string, id []byte) {
	var buf bytes.Buffer
	buf.Write(n.Theme.Owner().UUID())
	buf.Write(id)
	prefix := "note-"
	if subtype != "" {
		prefix = prefix + subtype + "-"
	}
	n.doc.ID = prefix + hex.EncodeToString(buf.Bytes())
}

func (n *Note) newNoteDoc() {
	n.doc = noteDoc{
		Comment:     &n.Comment,
		Attachments: make(map[string]*model.Attachment),
	}
}

func (n *Note) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.doc)
}

func (n *Note) UnmarshalJSON(data []byte) error {
	n.newNoteDoc()
	return json.Unmarshal(data, &n.doc)
}

func ImportAnkiNote(t *theme.Theme, note *anki.Note) (*Note, error) {
	now := time.Now()
	n := New(t, 0)
	n.setID("anki", note.AnkiID())
	n.doc.Created = nil
	n.doc.Modified = note.Modified
	n.doc.Imported = &now
	if err := n.Save(); err != nil {
		if !pouchdb.IsConflict(err) {
			return nil, err
		}
		existing, err2 := FetchNote(n.Theme, n.ID())
		if err2 != nil {
			return nil, fmt.Errorf("Fetching note: %s\n", err2)
		}
		if err := existing.MergeImport(n); err != nil {
			if model.NoChange(err) {
				return existing, nil
			}
			return nil, err
		}
		fmt.Printf("Doc: %v", existing)
		b, e := json.Marshal(existing)
		fmt.Printf("JSON err: %s\n", e)
		fmt.Printf("JSON: %s\n", b)
		if err := existing.Save(); err != nil {
			return nil, err
		} else {
			return existing, nil
		}
	}
	return n, nil
}

func (n *Note) Save() error {
	u, err := user.CurrentUser()
	if err != nil {
		return err
	}
	db := model.NewDB(u.DBName())
	if rev, err := db.Put(n); err != nil {
		return err
	} else {
		n.doc.Rev = rev
	}
	return nil
}

func FetchNote(t *theme.Theme, id string) (*Note, error) {
	u, err := user.CurrentUser()
	if err != nil {
		fmt.Printf("No current user\n")
		return nil, err
	}
	db := model.NewDB(u.DBName())
	n := &Note{}
	if err := db.Get(id, n, pouchdb.Options{}); err != nil {
		fmt.Printf("10: id  = %s\n", id)
		fmt.Printf("10: err = %s\n", err)
		return nil, err
	}
	n.Theme = t
	if err := n.check(); err != nil {
		return nil, err
	}
	return n, nil
}

func (n *Note) check() error {
	if !strings.HasPrefix(n.doc.ModelID, n.baseModelID()) {
		return errors.New("ModelID does not match expected")
	}
	return nil
}

func (n *Note) MergeImport(newNote *Note) error {
	if n.Imported() == nil {
		return errors.New("Conflict. Cannot MergeImport to a non-imported note")
	}
	if n.Modified().After(*newNote.Imported()) {
		return errors.New("The note has been modified since import. Merge not possible.")
	}
	if n.Modified().Equal(*newNote.Modified()) {
		return model.NewModelErrorNoChange()
	}
	n.Comment = newNote.Comment
	n.doc.Created = newNote.doc.Created
	n.doc.Modified = newNote.doc.Modified
	n.doc.Imported = newNote.doc.Imported
	n.doc.ModelID = newNote.doc.ModelID
	n.doc.Tags = newNote.doc.Tags
	n.doc.FieldValues = newNote.doc.FieldValues
	n.doc.Attachments = newNote.doc.Attachments
	return nil
}

// Read-only getters
func (n *Note) ID() string {
	return n.doc.ID
}

func (n *Note) Rev() string {
	return n.doc.Rev
}

func (n *Note) Created() *time.Time {
	if n.doc.Created == nil {
		return nil
	}
	ts := *n.doc.Created
	return &ts
}

func (n *Note) Modified() *time.Time {
	if n.doc.Modified == nil {
		return nil
	}
	ts := *n.doc.Modified
	return &ts
}

func (n *Note) Imported() *time.Time {
	if n.doc.Imported == nil {
		return nil
	}
	ts := *n.doc.Imported
	return &ts
}
