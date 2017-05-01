package repo

import (
	"context"

	"github.com/flimzy/log"
	"github.com/pkg/errors"
)

//go:generate go-bindata -pkg repo -nocompress -prefix files -o data.go files

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
	updated, err := Upsert(context.TODO(), db, ddoc, nil)
	if err != nil {
		log.Debugf("error creating view: %s\n", err)
	}
	if updated {
		// Query the views, so the indexes are created immediately
		rows, _ := db.Query(context.TODO(), "cards", "NewCardsMap", map[string]interface{}{"limit": 1})
		if rows != nil {
			rows.Close()
		}
		rows, _ = db.Query(context.TODO(), "cards", "OldCardsMap", map[string]interface{}{"limit": 1})
		if rows != nil {
			rows.Close()
		}
	}
	return err
}

func bundleDBInit(db *DB) error {
	return nil
}
