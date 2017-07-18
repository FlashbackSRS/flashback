package model

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"

	"github.com/pkg/errors"

	fb "github.com/FlashbackSRS/flashback-model"
)

type inputFile interface {
	Bytes() ([]byte, error)
}

// ImportFile imports a *.fbb file, as from an HTML form submission.
func (r *Repo) ImportFile(f inputFile) error {
	if _, err := r.CurrentUser(); err != nil {
		return err
	}
	b, err := f.Bytes()
	if err != nil {
		return err
	}
	z, err := gzip.NewReader(bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer func() { _ = z.Close() }()
	return nil
}

// Import imports a .fbb file and stores the content
func (r *Repo) Import(ctx context.Context, f io.Reader) error {
	if _, err := r.CurrentUser(); err != nil {
		return err
	}
	pkg := &fb.Package{}
	if err := json.NewDecoder(f).Decode(pkg); err != nil {
		return errors.Wrap(err, "Unable to decode JSON")
	}
	if err := pkg.Validate(); err != nil {
		return err
	}

	bundle := pkg.Bundle
	if err := r.SaveBundle(ctx, bundle); err != nil {
		return err
	}
	//
	// 	// From this point on, we plow through the errors
	// 	var errs *multierror.Error
	//
	// 	// Themes
	// 	log.Debugln("Saving themes")
	// 	themes := make([]interface{}, len(pkg.Themes))
	// 	for i, theme := range pkg.Themes {
	// 		themes[i] = theme
	// 	}
	// 	if themeErr := bulkInsert(context.TODO(), bdb, themes...); themeErr != nil {
	// 		errs = multierror.Append(errs, errors.Wrapf(themeErr, "failure saving themes"))
	// 	}
	//
	// 	// Notes
	// 	log.Debugln("Saving notes")
	// 	notes := make([]interface{}, len(pkg.Notes))
	// 	for i, note := range pkg.Notes {
	// 		notes[i] = note
	// 	}
	// 	if noteErr := bulkInsert(context.TODO(), bdb, notes...); noteErr != nil {
	// 		errs = multierror.Append(errs, errors.Wrapf(noteErr, "failure saving notes"))
	// 	}
	//
	// 	// Decks
	// 	log.Debugln("Saving decks")
	// 	decks := make([]interface{}, len(pkg.Decks))
	// 	for i, deck := range pkg.Decks {
	// 		decks[i] = deck
	// 	}
	// 	if deckErr := bulkInsert(context.TODO(), bdb, decks...); deckErr != nil {
	// 		errs = multierror.Append(errs, errors.Wrapf(deckErr, "failure saving decks"))
	// 	}
	//
	// 	// Cards
	// 	log.Debugln("Saving cards")
	// 	cardsi := make([]interface{}, len(cards))
	// 	for i, card := range cards {
	// 		cardsi[i] = card
	// 	}
	// 	if cardErr := bulkInsert(context.TODO(), udb, cardsi...); cardErr != nil {
	// 		errs = multierror.Append(errs, errors.Wrapf(cardErr, "failure saving cards"))
	// 	}
	// 	return errs.ErrorOrNil()
	return nil
}

//
// func bulkInsert(ctx context.Context, db *DB, docs ...interface{}) error {
// 	results, err := db.BulkDocs(ctx, docs)
// 	if err != nil {
// 		return err
// 	}
// 	var errs *multierror.Error
// 	for results.Next() {
// 		if err := results.UpdateErr(); err != nil {
// 			errs = multierror.Append(errs, errors.Wrapf(err, "failed to save doc %s", results.ID()))
// 		}
// 	}
// 	return errs.ErrorOrNil()
// }
