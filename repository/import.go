package repo

import (
	"context"
	"encoding/json"
	"io"

	"github.com/flimzy/log"
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
	bdb, err := user.BundleDB(bundle)
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

	cardMap := map[string]*fb.Card{}
	for _, c := range pkg.Cards {
		cardMap[c.Identity()] = c
	}

	cards := make([]*fb.Card, 0, len(cardMap))

	for _, d := range pkg.Decks {
		for _, id := range d.Cards.All() {
			c, ok := cardMap[id]
			if !ok {
				return errors.Errorf("Card '%s' listed in deck, but not found in package", id)
			}
			cards = append(cards, c)
		}
	}

	// From this point on, we plow through the errors
	var errs *multierror.Error

	// Themes
	log.Debugln("Saving themes")
	if err := bulkInsert(context.TODO(), bdb, pkg.Themes); err != nil {
		errs = multierror.Append(errs, errors.Wrapf(err, "failure saving themes"))
	}

	// Notes
	log.Debugln("Saving notes")
	if err := bulkInsert(context.TODO(), bdb, pkg.Notes); err != nil {
		errs = multierror.Append(errs, errors.Wrapf(err, "failure saving notes"))
	}

	// Decks
	log.Debugln("Saving decks")
	if err := bulkInsert(context.TODO(), bdb, pkg.Decks); err != nil {
		errs = multierror.Append(errs, errors.Wrapf(err, "failure saving decks"))
	}

	// Cards
	log.Debugln("Saving cards")
	if err := bulkInsert(context.TODO(), udb, cards); err != nil {
		errs = multierror.Append(errs, errors.Wrapf(err, "failure saving cards"))
	}
	return errs.ErrorOrNil()
}

func bulkInsert(ctx context.Context, db *DB, docs interface{}) error {
	results, err := db.BulkDocs(ctx, docs)
	if err != nil {
		return err
	}
	var errs *multierror.Error
	for results.Next() {
		if err := results.UpdateErr(); err != nil {
			errs = multierror.Append(errs, errors.Wrapf(err, "failed to save doc %s", results.ID()))
		}
	}
	return errs
}
