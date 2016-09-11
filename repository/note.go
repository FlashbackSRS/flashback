package repo

import (
	"fmt"

	"github.com/flimzy/log"

	"github.com/FlashbackSRS/flashback-model"
	pouchdb "github.com/flimzy/go-pouchdb"
	"github.com/pkg/errors"
)

// Note is a wrapper around a fb.Note object
type Note struct {
	*fb.Note
	db    *DB
	theme *Theme
	model *Model
}

// Theme returns the card's associated Theme
func (n *Note) Theme() (*Theme, error) {
	if err := n.fetchTheme(); err != nil {
		return nil, errors.Wrap(err, "Error fetching theme for Theme()")
	}
	return n.theme, nil

}

// Model returns the card's associated Model
func (n *Note) Model() (*Model, error) {
	if err := n.fetchTheme(); err != nil {
		return nil, errors.Wrap(err, "Error fetching theme for Model()")
	}
	return n.model, nil
}

func (n *Note) fetchTheme() error {
	if n.theme != nil {
		// Nothing to do
		return nil
	}
	log.Debugf("Fetching theme %s", n.ThemeID)
	t := &fb.Theme{}
	if err := n.db.Get(n.ThemeID, t, pouchdb.Options{Attachments: true}); err != nil {
		fmt.Printf("Error: %s\n", err)
		return errors.Wrapf(err, "fetchTheme() can't fetch theme-%s", n.ThemeID)
	}
	m := t.Models[n.ModelID]
	n.theme = &Theme{t}
	n.model = &Model{m}
	n.model.Theme = t
	n.SetModel(m)

	log.Debugf("Fetched this model: %v\n", n.model)
	log.Debugf("Fetched this theme: %v\n", n.theme)
	return nil
}
