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
	"github.com/FlashbackSRS/flashback/model/progress"
)

type inputFile interface {
	Bytes() ([]byte, error)
}

// ImportFile imports a *.fbb file, as from an HTML form submission.
func (r *Repo) ImportFile(ctx context.Context, f inputFile, reporter progress.ReportFunc) error {
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
	return r.Import(ctx, z, reporter)
}

// Import imports a .fbb file and stores the content
func (r *Repo) Import(ctx context.Context, f io.Reader, reporter progress.ReportFunc) error {
	if reporter == nil {
		reporter = func(_, _ uint64, _ float64) {}
	}
	prog := progress.New(reporter)
	steps := prog.NewComponent()
	bundleDocs := prog.NewComponent()
	userDocs := prog.NewComponent()
	steps.Total(4)
	steps.Increment(1)
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

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error)
	defer close(errCh)
	go func() {
		err := importBundleDocs(ctx, bdb, pkg, bundleDocs)
		if err != nil {
			cancel()
		}
		errCh <- err
	}()

	go func() {
		err := importUserDocs(ctx, udb, pkg, userDocs)
		if err != nil {
			cancel()
		}
		errCh <- err
	}()

	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			return err
		}
	}

	log.Printf("Imported:\n%d Bundles\n%d Themes\n%d Decks\n%d Notes\n%d Cards\n",
		1, len(pkg.Themes), len(pkg.Decks), len(pkg.Notes), len(pkg.Cards))
	return nil
}

const bulkBatchSize = 50

func importBundleDocs(ctx context.Context, db kivikDB, pkg *fb.Package, prog *progress.Component) error {
	docCount := len(pkg.Themes) + len(pkg.Notes) + len(pkg.Decks)
	prog.Total(uint64(docCount*2 + 1))
	prog.Increment(uint64(docCount)) // To account for the reading delay
	docs := make([]FlashbackDoc, 0, docCount)
	prog.Increment(1)
	for _, theme := range pkg.Themes {
		docs = append(docs, theme)
	}
	for _, note := range pkg.Notes {
		docs = append(docs, note)
	}
	for _, deck := range pkg.Decks {
		docs = append(docs, deck)
	}

	for len(docs) > 0 {
		batchSize := bulkBatchSize
		if batchSize > len(docs) {
			batchSize = len(docs)
		}
		if err := bulkInsert(ctx, db, docs[0:batchSize]...); err != nil {
			return err
		}
		prog.Increment(uint64(batchSize))
		docs = docs[batchSize:]
	}
	return nil
}

func importUserDocs(ctx context.Context, db kivikDB, pkg *fb.Package, prog *progress.Component) error {
	docCount := len(pkg.Cards)
	prog.Total(uint64(docCount*2 + 1))
	prog.Increment(uint64(docCount)) // To account for the reading delay
	docs := make([]FlashbackDoc, 0, docCount)
	prog.Increment(1)
	for _, card := range pkg.Cards {
		docs = append(docs, card)
	}

	for len(docs) > 0 {
		batchSize := bulkBatchSize
		if batchSize > len(docs) {
			batchSize = len(docs)
		}
		if err := bulkInsert(ctx, db, docs[0:batchSize]...); err != nil {
			return err
		}
		prog.Increment(uint64(batchSize))
		docs = docs[batchSize:]
	}

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
