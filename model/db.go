package model

import (
	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/go-pouchdb/plugins/find"
)

type DB struct {
	*pouchdb.PouchDB
	*find.PouchPluginFind
}

func NewDB(name string) *DB {
	db := pouchdb.New(name)
	return &DB{
		db,
		find.New(db),
	}
}

func (db *DB) Compact() error {
	return db.PouchDB.Compact(pouchdb.Options{})
}
