package repo

import (
	"encoding/json"
	"io"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback-model"
)

// Import imports a .fbb file and stores the content
func Import(user *User, r io.Reader) error {
	pkg := &fb.Package{}
	err := json.NewDecoder(r).Decode(pkg)
	if err != nil {
		return errors.Wrap(err, "Unable to decode JSON")
	}

	udb, err := user.DB()
	if err != nil {
		return errors.Wrap(err, "Unable to connect to User DB")
	}
	bundle := pkg.Bundle
	bdb, err := BundleDB(bundle)
	if err != nil {
		return errors.Wrap(err, "Unable to connect to Bundle DB")
	}

	if err := udb.Save(bundle); err != nil {
		return errors.Wrap(err, "Unable to save Bundle to User DB")
	}
	bundle.Rev = nil
	if e := bdb.Save(bundle); e != nil {
		return errors.Wrap(e, "Unable to save Bundle to Bundle DB")
	}

	cards := make([]*fb.Card, 0, 100)

	for _, d := range pkg.Decks {
		for _, id := range d.Cards.All() {
			c, err := fb.NewCard(id)
			if err != nil {
				return errors.Wrapf(err, "Unable to create card %s", id)
			}
			cards = append(cards, c)
		}
	}

	// From this point on, we plow through the errors
	var errs *multierror.Error

	// Themes
	for _, t := range pkg.Themes {
		if err := bdb.Save(t); err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "Unable to save Theme %s", t.ID.Identity()))
		}
	}

	// Notes
	for _, n := range pkg.Notes {
		if err := bdb.Save(n); err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "Unable to save Note %s", n.ID.Identity()))
		}
	}

	// Decks
	for _, d := range pkg.Decks {
		if err := bdb.Save(d); err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "Unable to save Deck %s", d.ID.Identity()))
			continue
		}
	}

	// Cards
	for _, c := range cards {
		if err := udb.Save(c); err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "Unable to save Card %s", c.Identity()))
		}
	}
	return errs.ErrorOrNil()
}
