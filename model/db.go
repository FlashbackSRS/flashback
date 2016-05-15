package model

import (
	"errors"

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

// Not sure where the below stuff should live, so I'm just sticking it here for now.
type ModelError struct {
	e        error
	noChange bool
}

func (me *ModelError) Error() string {
	return me.e.Error()
}

func NewModelError(err error) *ModelError {
	return &ModelError{
		e: err,
	}
}

func NewModelErrorNoChange() *ModelError {
	return &ModelError{
		e:        errors.New("No change"),
		noChange: true,
	}
}

// ErrorMessage returns the message portion of a PouchError, or "" for other errors
func NoChange(err error) bool {
	switch me := err.(type) {
	case *ModelError:
		return me.noChange
	}
	return false
}
