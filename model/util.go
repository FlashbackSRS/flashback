package model

import (
	"context"
	"encoding/json"
	"reflect"
	"time"

	"github.com/flimzy/kivik"
	"github.com/pkg/errors"
)

// FlashbackDoc is a generic interface for all types of FB docs
type FlashbackDoc interface {
	DocID() string
	SetRev(string)
	ImportedTime() *time.Time
	ModifiedTime() *time.Time
	MergeImport(interface{}) (bool, error)
	json.Marshaler
	json.Unmarshaler
}

type saveDB interface {
	Put(context.Context, string, interface{}) (string, error)
	Get(context.Context, string, ...kivik.Options) (*kivik.Row, error)
}

func saveDoc(ctx context.Context, db saveDB, doc FlashbackDoc) error {
	var rev string
	var err error
	if rev, err = db.Put(context.TODO(), doc.DocID(), doc); err != nil {
		if kivik.StatusCode(err) != kivik.StatusConflict ||
			// Don't attempt to merge a non-import
			doc.ImportedTime() == nil {
			return err
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
		if imported == nil {
			return err
		}
		if existing.ModifiedTime().After(*imported) {
			// The existing document was modified after import, so we won't allow re-importing
			return err
		}
		var changed bool
		if changed, e = doc.MergeImport(existing); e != nil {
			return errors.Wrap(e, "failed to merge into existing document")
		}
		if changed {
			if rev, e = db.Put(context.TODO(), doc.DocID(), doc); e != nil {
				return errors.Wrap(e, "failed to store updated document")
			}
		} else {
			existingValue := reflect.ValueOf(&existing).Elem()
			reflect.ValueOf(&doc).Elem().Set(existingValue)
			return nil
		}
	}
	doc.SetRev(rev)
	return nil
}
