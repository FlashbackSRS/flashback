package bundle

import (
	"encoding/json"
	"errors"
	"math/rand"
	"time"

	"github.com/flimzy/flashback/anki"
	"github.com/flimzy/flashback/model/doc"
	"github.com/flimzy/flashback/model/user"
)

type bundleDoc struct {
	ID           string     `json:"_id"`
	Rev          string     `json:"_rev,omitempty"`
	Type         string     `json:"type"`
	Created      *time.Time `json:"created,omitempty"`
	Imported     *time.Time `json:"imported,omitempty"`
	Modified     *time.Time `json:"modified,omitempty"`
	ImportedFrom string     `json:"importedfrom,omitempty"`
	Name         string     `json:"name,omitempty"`
	Description  string     `json:"description,omitempty"`
	Owner        string     `json:"owner"`
}

type Bundle struct {
	doc   *bundleDoc
	owner *user.User
	// Permissions
}

func (b *Bundle) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.doc)
}

func (b *Bundle) UnmarshalJSON(data []byte) error {
	b.newInternalDoc()
	return json.Unmarshal(data, &b.doc)
}

type NewBundle struct {
	Owner        *user.User
	ImportedFrom string
	Key          []byte
}

func New(tmpl *NewBundle) (*Bundle, error) {
	if tmpl.Owner == nil {
		return nil, errors.New("Must specify owner")
	}
	b := Bundle{
		owner: tmpl.Owner,
	}
	b.newInternaldoc()
	now := time.Now()
	b.doc.Created = &now
	b.doc.Modified = &now
	b.doc.ImportedFrom = tmpl.ImportedFrom
	key := tmpl.Key
	if len(key) == 0 {
		var r [24]byte
		rand.Read(r)
		key = r[:]
	}
	b.setID(key)
	return &b
}

func (b *Bundle) newInternalDoc() {
	b.doc = bundleDoc{
		Type:        "bundle",
		Name:        &b.Name,
		Description: &b.Description,
	}
	if b.owner != nil {
		b.doc.Owner = b.owner.UUID().String()
	}
}

func (b *Bundle) setID(key []byte) {
	h := sha1.New()
	h.Write(b.owner.UUID())
	if b.ImportedFrom != "" {
		h.Write([]byte("-" + b.ImportedFrom + "-"))
	}
	h.Write(key)
	b.doc.ID = h.Sum(nil)
}

func CreateAnkiBundle(c *anki.Collection) (*Bundle, error) {
	u, err := user.CurrentUser()
	if err != nil {
		return nil, err
	}
	key, err := c.Created.MarshalText()
	if err != nil {
		return nil, err
	}
	b := New(NewBundle{
		Owner:        u,
		ImportedFrom: "anki",
		Key:          key,
	})
	b.doc.Created = c.Created
	b.doc.Modified = c.Modified
	now = time.Now()
	b.doc.Imported = &now
	b.Name = "Imported Anki Collection"
	// Save() -- copy from Theme
}
