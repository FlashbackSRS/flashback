package model

import (
	"context"
	"net/http"
	"testing"

	"github.com/flimzy/kivik"
)

func TestWrapDB(t *testing.T) {
	db := &kivik.DB{}
	wdb := wrapDB(db)
	if db != wdb.(*dbWrapper).DB {
		t.Errorf("Unexpected result")
	}
}

func TestWrappedGet(t *testing.T) {
	expected := http.StatusNotFound
	db := testDB(t)
	_, err := db.Get(context.Background(), "foo")
	status := kivik.StatusCode(err)
	if status != expected {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestWrappedQuery(t *testing.T) {
	db := testDB(t)
	_, err := db.Query(context.Background(), "", "")
	status := kivik.StatusCode(err)
	if status != http.StatusNotFound && // for PouchDB
		status != http.StatusNotImplemented { // for memory db
		t.Errorf("Unxpected error: %s", err)
	}
}
