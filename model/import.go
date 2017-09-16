package model

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"

	"github.com/FlashbackSRS/flashback/model/progress"
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
	return r.Import(ctx, z, nil)
}

// Import imports a .fbb file and stores the content
func (r *Repo) Import(ctx context.Context, f io.Reader, reporter progress.ReportFunc) error {
	if reporter == nil {
		reporter = func(_, _ uint64, _ float64) {}
	}
	prog := progress.New(reporter)
	steps := prog.NewComponent()
	bundleDocs := prog.NewComponent()
	cardDocs := prog.NewComponent()
	steps.Total(5)
	udb, err := r.userDB(ctx)
	if err != nil {
		return err
	}
	pkg := &fb.Package{}
	if e := json.NewDecoder(f).Decode(pkg); e != nil {
		return errors.Wrap(e, "Unable to decode JSON")
	}
	steps.Increment(1)
	if e := pkg.Validate(); e != nil {
		return e
	}

	bundle := pkg.Bundle
	bundle.Owner = r.user
	if e := r.SaveBundle(ctx, bundle); e != nil {
		return e
	}
	steps.Increment(1)

	bdb, err := r.bundleDB(ctx, bundle)
	if err != nil {
		return err
	}
	steps.Increment(1)

	if err := importBundleDocs(ctx, bdb, pkg, bundleDocs); err != nil {
		return err
	}
	steps.Increment(1)

	if err := importCardDocs(ctx, udb, pkg, cardDocs); err != nil {
		return err
	}
	steps.Increment(1)

	log.Printf("Imported:\n%d Bundles\n%d Themes\n%d Decks\n%d Notes\n%d Cards\n",
		1, len(pkg.Themes), len(pkg.Decks), len(pkg.Notes), len(pkg.Cards))
	return nil
}

func importBundleDocs(ctx context.Context, db kivikDB, pkg *fb.Package, prog *progress.Component) error {
	docs := make([]FlashbackDoc, 0, len(pkg.Themes)+len(pkg.Notes)+len(pkg.Decks))
	prog.Total(uint64(cap(docs)))
	defer prog.Progress(uint64(cap(docs)))
	for _, theme := range pkg.Themes {
		docs = append(docs, theme)
	}
	for _, note := range pkg.Notes {
		docs = append(docs, note)
	}
	for _, deck := range pkg.Decks {
		docs = append(docs, deck)
	}
	return bulkInsert(ctx, db, docs...)
}

func importCardDocs(ctx context.Context, db kivikDB, pkg *fb.Package, prog *progress.Component) error {

	cards := make([]FlashbackDoc, 0, len(pkg.Cards))
	prog.Total(uint64(cap(cards)))
	defer prog.Progress(uint64(cap(cards)))
	for _, card := range pkg.Cards {
		cards = append(cards, card)
	}

	return bulkInsert(ctx, db, cards...)
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
