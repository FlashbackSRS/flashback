package model

import (
	"context"

	"github.com/flimzy/kivik"
)

type querier interface {
	Query(ctx context.Context, ddoc, view string, options ...kivik.Options) (kivikRows, error)
}

type putter interface {
	Put(context.Context, string, interface{}) (string, error)
}

type getter interface {
	Get(ctx context.Context, docID string, options ...kivik.Options) (kivikRow, error)
}

type bulkDocer interface {
	BulkDocs(context.Context, interface{}) (*kivik.BulkResults, error)
}

type getPutter interface {
	putter
	getter
}

type getPutBulkDocer interface {
	getter
	putter
	bulkDocer
}

type kivikDB interface {
	getter
	putter
	querier
	bulkDocer
}

type dbWrapper struct {
	*kivik.DB
}

var _ kivikDB = &dbWrapper{}

func (db *dbWrapper) Get(ctx context.Context, docID string, options ...kivik.Options) (kivikRow, error) {
	return db.DB.Get(ctx, docID, options...)
}

func (db *dbWrapper) Query(ctx context.Context, ddoc, view string, options ...kivik.Options) (kivikRows, error) {
	return db.DB.Query(ctx, ddoc, view, options...)
}

func wrapDB(db *kivik.DB) kivikDB {
	return &dbWrapper{DB: db}
}

type kivikRow interface {
	ScanDoc(dest interface{}) error
}

type kivikRows interface {
	Close() error
	Next() bool
	ScanDoc(dest interface{}) error
	TotalRows() int64
}
