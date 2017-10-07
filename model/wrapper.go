package model

import (
	"context"

	"github.com/flimzy/kivik"
)

type replicator interface {
	Replicate(ctx context.Context, targetDSN, sourceDSN string, options ...kivik.Options) (*kivik.Replication, error)
}

type dsner interface {
	DSN() string
}

type kivikClient interface {
	CreateDB(ctx context.Context, dbName string, options ...kivik.Options) error
	DB(ctx context.Context, dbName string, options ...kivik.Options) (kivikDB, error)
	DestroyDB(ctx context.Context, dbName string, options ...kivik.Options) error
	dsner
	replicator
}

type clientWrapper struct {
	*kivik.Client
}

var _ kivikClient = &clientWrapper{}

func (c *clientWrapper) DB(ctx context.Context, dbName string, options ...kivik.Options) (kivikDB, error) {
	db, err := c.Client.DB(ctx, dbName, options...)
	if err != nil {
		return nil, err
	}
	return wrapDB(db), nil
}

func wrapClient(c *kivik.Client) kivikClient {
	return &clientWrapper{Client: c}
}

type queryGetter interface {
	querier
	getter
}

type querier interface {
	Query(ctx context.Context, ddoc, view string, options ...kivik.Options) (kivikRows, error)
}

type putter interface {
	Put(context.Context, string, interface{}) (string, error)
}

type getter interface {
	Get(ctx context.Context, docID string, options ...kivik.Options) (kivikRow, error)
}

type allDocer interface {
	AllDocs(ctx context.Context, options ...kivik.Options) (kivikRows, error)
}

type bulkDocer interface {
	BulkDocs(context.Context, interface{}) (kivikBulkResults, error)
}

type kivikBulkResults interface {
	Close() error
	Next() bool
	Err() error
	ID() string
	UpdateErr() error
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

type attachmentGetter interface {
	GetAttachment(ctx context.Context, docID, rev, filename string) (*kivik.Attachment, error)
}

type finder interface {
	Find(ctx context.Context, query interface{}) (kivikRows, error)
}

type deleter interface {
	Delete(ctx context.Context, docID, rev string) (newRev string, err error)
}

type statser interface {
	Stats(ctx context.Context) (*kivik.DBStats, error)
}

type clientNamer interface {
	Client() kivikClient
	Name() string
}

type kivikDB interface {
	getter
	putter
	querier
	bulkDocer
	finder
	deleter
	statser
	clientNamer
	attachmentGetter
	allDocer
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

func (db *dbWrapper) BulkDocs(ctx context.Context, docs interface{}) (kivikBulkResults, error) {
	return db.DB.BulkDocs(ctx, docs)
}

func (db *dbWrapper) Find(ctx context.Context, query interface{}) (kivikRows, error) {
	return db.DB.Find(ctx, query)
}

func (db *dbWrapper) AllDocs(ctx context.Context, options ...kivik.Options) (kivikRows, error) {
	return db.DB.AllDocs(ctx, options...)
}

func (db *dbWrapper) Client() kivikClient {
	return wrapClient(db.DB.Client())
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
	ScanValue(dest interface{}) error
	ScanKey(dest interface{}) error
	TotalRows() int64
	ID() string
	Err() error
	Key() string
}
