package model

import (
	"context"
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
	db := testDB(t)
	_, err := db.Get(context.Background(), "foo")
	checkErr(t, "missing", err)
}

func TestWrappedQuery(t *testing.T) {
	db := testDB(t)
	_, err := db.Query(context.Background(), "", "")
	checkErr(t, "kivik: not yet implemented in memory driver", err)
}
