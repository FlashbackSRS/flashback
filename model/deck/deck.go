package deck

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/flimzy/go-pouchdb"

	"github.com/flimzy/flashback/anki"
	"github.com/flimzy/flashback/model"
	"github.com/flimzy/flashback/model/user"
)

type deckDoc struct {
	ID          string     `json:"_id"`
	Rev         string     `json:"_rev,omitempty"`
	Type        string     `json:"type"`
	Name        *string    `json:"name"`
	Owner       string     `json:"owner"`
	Description *string    `json:"description,omitempty"`
	Created     *time.Time `json:"created,omitempty"`
	Modified    *time.Time `json:"modified,omitempty"`
	Imported    *time.Time `json:"imported,omitempty"`
}

func (d *Deck) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.doc)
}

func (d *Deck) UnmarshalJSON(data []byte) error {
	d.newDeckDoc()
	return json.Unmarshal(data, &d.doc)
}

type Deck struct {
	doc         deckDoc
	owner       *user.User
	Name        string
	Description string
}

func New(owner *user.User) *Deck {
	d := Deck{
		owner: owner,
	}
	d.newDeckDoc()
	now := time.Now()
	d.doc.Created = &now
	d.doc.Modified = &now
	return &d
}

func (d *Deck) setID(subtype string, id []byte) {
	var buf bytes.Buffer
	buf.Write(d.owner.UUID())
	buf.Write(id)
	prefix := "deck-"
	if subtype != "" {
		prefix = prefix + subtype + "-"
	}
	d.doc.ID = prefix + hex.EncodeToString(buf.Bytes())
}

func (d *Deck) newDeckDoc() {
	d.doc = deckDoc{
		Type:        "deck",
		Name:        &d.Name,
		Description: &d.Description,
	}
	if d.owner != nil {
		d.doc.Owner = d.owner.UUID().String()
	}
}

func ImportAnkiDeck(deck *anki.Deck) (*Deck, error) {
	u, err := user.CurrentUser()
	if err != nil {
		return nil, err
	}
	d := New(u)
	d.setID("anki", deck.AnkiID())
	d.Name = deck.Name
	if deck.ID == 1 && deck.Name == "Default" {
		d.Name = "Default Anki Deck"
	}
	d.doc.Created = nil
	d.doc.Modified = deck.Modified
	now := time.Now()
	d.doc.Imported = &now
	if err := d.Save(); err != nil {
		if !pouchdb.IsConflict(err) {
			return nil, err
		}
		existing, err2 := FetchDeck(d.owner, d.ID())
		if err2 != nil {
			return nil, fmt.Errorf("Fetching deck: %s\n", err2)
		}
		if err := existing.MergeImport(d); err != nil {
			if model.NoChange(err) {
				return existing, nil
			}
			return nil, err
		}
		if err := existing.Save(); err != nil {
			return nil, err
		} else {
			return existing, nil
		}
	}
	return d, nil
}

func (d *Deck) Save() error {
	u, err := user.CurrentUser()
	if err != nil {
		return err
	}
	db := model.NewDB(d.ID())
	if rev, err := db.Put(d); err != nil {
		return err
	} else {
		d.doc.Rev = rev
	}
	udb := model.NewDB(u.DBName())
	s := d.stub()
	if _, err := udb.Put(s); err != nil && !pouchdb.IsConflict(err) {
		return err
	}
	return nil
}

func (d *Deck) stub() *model.Stub {
	return &model.Stub{
		ID:         d.doc.ID,
		Type:       "stub",
		ParentType: "deck",
	}
}

func FetchDeck(u *user.User, id string) (*Deck, error) {
	db := model.NewDB(id)
	d := &Deck{}
	if err := db.Get(id, d, pouchdb.Options{}); err != nil {
		return nil, err
	}
	d.owner = u
	if err := d.check(); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *Deck) check() error {
	if d.doc.Owner != d.owner.ID() {
		return errors.New("Deck owner does not match expected")
	}
	return nil
}

func (d *Deck) Created() *time.Time {
	if d.doc.Created == nil {
		return nil
	}
	ts := *d.doc.Created
	return &ts
}

func (d *Deck) Modified() *time.Time {
	if d.doc.Modified == nil {
		return nil
	}
	ts := *d.doc.Modified
	return &ts
}

func (d *Deck) Imported() *time.Time {
	if d.doc.Imported == nil {
		return nil
	}
	ts := *d.doc.Imported
	return &ts
}

func (d *Deck) MergeImport(n *Deck) error {
	if d.Imported() == nil {
		return errors.New("Conflict. Cannot MergeImport to a non-imported deck")
	}
	if d.Modified().After(*n.Imported()) {
		return errors.New("The deck has been modified since import. Merge not possible")
	}
	d.Name = n.Name
	d.Description = n.Description
	d.doc.Owner = n.doc.Owner
	d.doc.Created = n.doc.Created
	d.doc.Modified = n.doc.Modified
	d.doc.Imported = n.doc.Imported
	return nil
}

// Read-only getters
func (d *Deck) ID() string {
	return d.doc.ID
}
