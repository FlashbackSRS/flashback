package repo

import (
	pouchdb "github.com/flimzy/go-pouchdb"
	"github.com/flimzy/log"
	"github.com/pkg/errors"
)

func userDBInit(db *DB) error {
	newCardsMapFunc, err := Asset("NewCardsMap.js")
	if err != nil {
		return errors.Wrap(err, "failed to get NewCardsMap.js")
	}
	oldCardsMapFunc, err := Asset("OldCardsMap.js")
	if err != nil {
		return errors.Wrap(err, "failed to get OldCardsMap.js")
	}
	ddoc := Map{
		"_id": "_design/cards",
		"views": Map{
			"NewCardsMap": Map{
				"map": string(newCardsMapFunc),
			},
			"OldCardsMap": Map{
				"map": string(oldCardsMapFunc),
			},
		},
	}
	log.Debugf("Creating _design/cards\n")
	updated, err := Upsert(db, ddoc, pouchdb.Options{})
	if err != nil {
		log.Debugf("error creating view: %s\n", err)
	}
	if updated {
		// Query the views, so the indexes are created immediately
		var result interface{}
		_ = db.Query("cards/NewCardsMap", &result, pouchdb.Options{Limit: 1})
		_ = db.Query("cards/OldCardsMap", &result, pouchdb.Options{Limit: 1})
	}
	return err
}

func bundleDBInit(db *DB) error {
	return nil
}
