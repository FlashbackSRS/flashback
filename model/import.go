package model

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"

	"github.com/flimzy/kivik"
	"github.com/flimzy/log"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	fb "github.com/FlashbackSRS/flashback-model"
)

type inputFile interface {
	Bytes() ([]byte, error)
}

// ImportFile imports a *.fbb file, as from an HTML form submission.
func (r *Repo) ImportFile(ctx context.Context, f inputFile) error {
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
	return r.Import(ctx, z)
}

// Import imports a .fbb file and stores the content
func (r *Repo) Import(ctx context.Context, f io.Reader) error {
	udb, err := r.userDB(ctx)
	if err != nil {
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
	bundle.Owner = r.user
	if err := r.SaveBundle(ctx, bundle); err != nil {
		return err
	}

	bdb, err := r.bundleDB(ctx, bundle)
	if err != nil {
		return err
	}

	docs := make([]FlashbackDoc, 0, len(pkg.Themes)+len(pkg.Notes)+len(pkg.Decks))
	for _, theme := range pkg.Themes {
		docs = append(docs, theme)
	}
	for _, note := range pkg.Notes {
		docs = append(docs, note)
	}
	for _, deck := range pkg.Decks {
		docs = append(docs, deck)
	}
	if err := bulkInsert(ctx, bdb, docs...); err != nil {
		return err
	}

	cards := make([]FlashbackDoc, 0, len(pkg.Cards))
	for _, card := range pkg.Cards {
		cards = append(cards, card)
	}

	if err := bulkInsert(ctx, udb, cards...); err != nil {
		return err
	}
	log.Printf("Imported:\n%d Bundles\n%d Themes\n%d Decks\n%d Notes\n%d Cards\n",
		1, len(pkg.Themes), len(pkg.Decks), len(pkg.Notes), len(pkg.Cards))
	return nil
}

func bulkInsert(ctx context.Context, db getPutBulkDocer, docs ...FlashbackDoc) error {
	results, err := db.BulkDocs(ctx, docs)
	if err != nil {
		return err
	}
	var errs *multierror.Error
	for i := 0; results.Next(); i++ {
		if err := results.UpdateErr(); err != nil {
			if kivik.StatusCode(err) == kivik.StatusConflict {
				if e := mergeDoc(ctx, db, docs[i]); e != nil {
					err = e
				} else {
					continue
				}
			}
			errs = multierror.Append(errs, errors.Wrapf(err, "failed to save doc %s", results.ID()))
		}
	}
	return errs.ErrorOrNil()
}
