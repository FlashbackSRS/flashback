package repo

import (
	"context"
	"reflect"

	"github.com/flimzy/kivik"
)

// Upsert creates or replaces the target document
func Upsert(ctx context.Context, db *DB, doc, opts map[string]interface{}) (bool, error) {
	_, err := db.Put(ctx, doc["_id"].(string), doc)
	if err == nil {
		return true, nil
	}
	if kivik.StatusCode(err) != kivik.StatusConflict {
		return false, err
	}
	var existing map[string]interface{}
	row, err := db.Get(ctx, doc["_id"].(string), opts)
	if err != nil {
		return false, err
	}
	if err = row.ScanDoc(&existing); err != nil {
		return false, err
	}
	doc["_rev"] = existing["_rev"]
	if reflect.DeepEqual(doc, existing) {
		// No update needed
		return false, nil
	}
	_, err = db.Put(ctx, doc["_id"].(string), doc)
	return true, err
}
