package model

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

// FlashbackDoc is a generic interface for all types of FB docs
type FlashbackDoc interface {
	DocID() string
	SetRev(string)
	ImportedTime() time.Time
	ModifiedTime() time.Time
	MergeImport(interface{}) (bool, error)
	json.Marshaler
	json.Unmarshaler
}

func mergeDoc(ctx context.Context, db getPutter, doc FlashbackDoc) error {
	// Don't attempt to merge a non-import
	if doc.ImportedTime().IsZero() {
		return errors.Status(kivik.StatusConflict, "document update conflict")
	}
	existing := reflect.New(reflect.TypeOf(doc).Elem()).Interface().(FlashbackDoc)
	row, e := db.Get(context.TODO(), doc.DocID())
	if e != nil {
		return errors.Wrap(e, "failed to fetch existing document")
	}
	if e = row.ScanDoc(&existing); e != nil {
		return errors.Wrap(e, "failed to parse existing document")
	}
	imported := existing.ImportedTime()
	if imported.IsZero() {
		return errors.Status(kivik.StatusConflict, "document update conflict")
	}
	if existing.ModifiedTime().After(imported) {
		// The existing document was modified after import, so we won't allow re-importing
		return errors.Status(kivik.StatusConflict, "document update conflict")
	}
	var changed bool
	if changed, e = doc.MergeImport(existing); e != nil {
		return errors.Wrap(e, "failed to merge into existing document")
	}
	if changed {
		rev, e := db.Put(context.TODO(), doc.DocID(), doc)
		if e != nil {
			return errors.Wrap(e, "failed to store updated document")
		}
		doc.SetRev(rev)
		return nil
	}
	existingValue := reflect.ValueOf(&existing).Elem()
	reflect.ValueOf(&doc).Elem().Set(existingValue)
	return nil
}

func saveDoc(ctx context.Context, db getPutter, doc FlashbackDoc) error {
	var rev string
	var err error
	if rev, err = db.Put(context.TODO(), doc.DocID(), doc); err != nil {
		if kivik.StatusCode(err) == kivik.StatusConflict {
			return mergeDoc(ctx, db, doc)
		}
		return err
	}
	doc.SetRev(rev)
	return nil
}

func getDoc(ctx context.Context, db getter, id string, dst interface{}) error {
	row, err := db.Get(ctx, id)
	if err != nil {
		return err
	}
	return row.ScanDoc(dst)
}

func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
