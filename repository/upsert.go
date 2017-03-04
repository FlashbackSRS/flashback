package repo

import (
	"reflect"

	pouchdb "github.com/flimzy/go-pouchdb"
)

// Upsert creates or replaces the target document
func Upsert(db *DB, doc map[string]interface{}, opts pouchdb.Options) (bool, error) {
	_, err := db.Put(doc)
	if err == nil {
		return true, nil
	}
	if !pouchdb.IsConflict(err) {
		return false, err
	}
	var existing map[string]interface{}
	err = db.Get(doc["_id"].(string), &existing, opts)
	if err != nil {
		return false, err
	}
	doc["_rev"] = existing["_rev"]
	if reflect.DeepEqual(doc, existing) {
		// No update needed
		return false, nil
	}
	_, err = db.Put(doc)
	return true, err
}
