package model

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

type mockAttachmentGetter struct {
	attachments map[string]*kivik.Attachment
	err         error
}

var _ attachmentGetter = &mockAttachmentGetter{}

func (db *mockAttachmentGetter) GetAttachment(_ context.Context, _, _, filename string) (*kivik.Attachment, error) {
	if db.err != nil {
		return nil, db.err
	}
	if att, ok := db.attachments[filename]; ok {
		return att, nil
	}
	return nil, errors.Status(kivik.StatusNotFound, "not found")
}

type mockAllDocer struct {
	kivikDB
	rows kivikRows
	err  error
}

var _ allDocer = &mockAllDocer{}

func (db *mockAllDocer) AllDocs(_ context.Context, options ...kivik.Options) (kivikRows, error) {
	return db.rows, db.err
}

type mockRows struct {
	rows, values, keys []string
	i, limit           int
	err                error
}

var _ kivikRows = &mockRows{}

func (r *mockRows) Err() error   { return r.err }
func (r *mockRows) Close() error { return nil }
func (r *mockRows) Next() bool {
	r.i++
	if r.limit > 0 {
		return r.i-1 < r.limit && r.i-1 < len(r.rows)
	}
	return r.i-1 < len(r.rows)
}
func (r *mockRows) TotalRows() int64 { return int64(len(r.rows)) }
func (r *mockRows) ID() string {
	doc := struct {
		ID string `json:"_id"`
	}{}
	_ = json.Unmarshal([]byte(r.rows[r.i-1]), &doc)
	return doc.ID
}
func (r *mockRows) scan(d interface{}, src []string) error {
	if r.limit > 0 && r.i-1 > r.limit {
		return io.EOF
	}
	if r.i-1 >= len(src) {
		return io.EOF
	}
	if err := json.Unmarshal([]byte(src[r.i-1]), &d); err != nil {
		return err
	}
	return nil
}
func (r *mockRows) ScanDoc(d interface{}) error   { return r.scan(d, r.rows) }
func (r *mockRows) ScanValue(d interface{}) error { return r.scan(d, r.values) }
func (r *mockRows) ScanKey(d interface{}) error   { return r.scan(d, r.keys) }

type mockBulkDocer struct {
	results kivikBulkResults
	err     error
}

var _ bulkDocer = &mockBulkDocer{}

func (db *mockBulkDocer) BulkDocs(_ context.Context, docs interface{}) (kivikBulkResults, error) {
	return db.results, db.err
}

type mockBulkResults struct {
	i    int
	errs []error
	err  error
}

var _ kivikBulkResults = &mockBulkResults{}

func (r *mockBulkResults) Close() error {
	r.errs = nil
	return nil
}

func (r *mockBulkResults) Err() error { return r.err }
func (r *mockBulkResults) ID() string { panic("not done") }
func (r *mockBulkResults) Next() bool {
	if r.i >= len(r.errs) {
		return false
	}
	r.i++
	return true
}
func (r *mockBulkResults) UpdateErr() error { return r.errs[r.i-1] }

type mockClient struct {
	kivikClient
	db  kivikDB
	err error
}

var _ kivikClient = &mockClient{}

func (c *mockClient) DB(_ context.Context, _ string, _ ...kivik.Options) (kivikDB, error) {
	return c.db, c.err
}

type mockQueryGetter struct {
	*mockQuerier
	row kivikRow
	err error
}

var _ queryGetter = &mockQueryGetter{}

func (db *mockQueryGetter) Get(_ context.Context, docID string, _ ...kivik.Options) (kivikRow, error) {
	return db.row, db.err
}

type mockRow string

var _ kivikRow = mockRow("")

func (r mockRow) ScanDoc(i interface{}) error {
	return json.Unmarshal([]byte(r), &i)
}

type mockQuerier struct {
	kivikDB
	options []kivik.Options
	rows    []*mockRows
	err     error
}

var _ querier = &mockQuerier{}

func optsEqual(opt, req kivik.Options) bool {
	for k, v := range opt {
		reqVal, ok := req[k]
		if !ok || !reflect.DeepEqual(v, reqVal) {
			return false
		}
	}
	return true
}

func (db *mockQuerier) Query(ctx context.Context, ddoc, view string, options ...kivik.Options) (kivikRows, error) {
	if db.err != nil {
		return nil, db.err
	}
	queryIndex := -1
	if len(db.rows) > 1 || len(db.options) > 0 {
		for i, opts := range db.options {
			if optsEqual(opts, options[0]) {
				queryIndex = i
				break
			}
		}
		if queryIndex < 0 {
			var d string
			if len(db.options) == 1 {
				d = diff.Interface(db.options[0], options[0]).String()
			} else {
				js, _ := json.MarshalIndent(options[0], "", "    ")
				d = string(js)
			}
			return nil, fmt.Errorf("Matching query not found in mock result set\n%s", d)
		}
	} else {
		queryIndex = 0
	}
	limit, _ := options[0]["limit"].(int)
	offset, _ := options[0]["skip"].(int)
	rows := db.rows[queryIndex]
	rows.limit = limit + offset
	rows.i = offset
	return rows, nil
}
