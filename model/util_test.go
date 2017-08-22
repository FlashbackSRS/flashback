package model

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

type testDoc struct {
	ID       string    `json:"_id"`
	Rev      string    `json:"_rev,omitempty"`
	ITime    time.Time `json:"imported_time,omitempty"`
	MTime    time.Time `json:"modified_time,omitempty"`
	Value    string    `json:"value,omitempty"`
	doMerge  bool
	mergeErr error
}

var _ FlashbackDoc = &testDoc{}

func (d *testDoc) DocID() string                { return d.ID }
func (d *testDoc) SetRev(rev string)            { d.Rev = rev }
func (d *testDoc) ImportedTime() time.Time      { return d.ITime }
func (d *testDoc) ModifiedTime() time.Time      { return d.MTime }
func (d *testDoc) MarshalJSON() ([]byte, error) { return json.Marshal(*d) }
func (d *testDoc) MergeImport(doc interface{}) (bool, error) {
	if d.doMerge {
		if newDoc, ok := doc.(*testDoc); ok {
			d.Rev = newDoc.Rev
			d.Value = "new value"
		}
		return true, nil
	}
	return false, d.mergeErr
}
func (d *testDoc) UnmarshalJSON(data []byte) error {
	type newDoc testDoc
	var doc newDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return err
	}
	*d = testDoc(doc)
	return nil
}

type mockDB struct {
	puts   int
	putErr []error
	getErr error
}

var _ getPutter = &mockDB{}

func (db *mockDB) Get(_ context.Context, _ string, _ ...kivik.Options) (kivikRow, error) {
	return nil, db.getErr
}

func (db *mockDB) Put(_ context.Context, _ string, _ interface{}) (string, error) {
	if len(db.putErr) < db.puts {
		return "", nil
	}
	err := db.putErr[db.puts]
	db.puts++
	return "", err
}

func TestSaveDoc(t *testing.T) {
	type sdTest struct {
		name     string
		db       getPutter
		doc      FlashbackDoc
		expected map[string]interface{}
		err      string
	}
	now := time.Now()
	then := time.Now().Add(-24 * time.Hour)
	tests := []sdTest{
		{
			name: "New doc",
			db:   testDB(t),
			doc:  &testDoc{ID: "foo"},
			expected: map[string]interface{}{
				"_id":           "foo",
				"_rev":          "1",
				"imported_time": "0001-01-01T00:00:00Z",
				"modified_time": "0001-01-01T00:00:00Z",
			},
		},
		{
			name: "Put Error",
			db:   &mockDB{putErr: []error{errors.New("put error")}},
			doc:  &testDoc{ID: "_foo"},
			err:  "put error",
		},
		{
			name: "Get Error",
			db: &mockDB{
				putErr: []error{errors.Status(http.StatusConflict, "conflict")},
				getErr: errors.New("get error"),
			},
			doc: &testDoc{ID: "_foo", ITime: now},
			err: "failed to fetch existing document: get error",
		},
		{
			name: "Conflict with non-imported doc",
			db: func() kivikDB {
				db := testDB(t)
				doc := map[string]string{
					"_id": "foo",
				}
				if _, err := db.Put(context.Background(), "foo", doc); err != nil {
					t.Fatal(err)
				}
				return db
			}(),
			doc: &testDoc{ID: "foo", ITime: now},
			err: "document update conflict",
		},
		{
			name: "Modified Doc Exists",
			db: func() kivikDB {
				db := testDB(t)
				doc := map[string]interface{}{
					"_id":           "foo",
					"imported_time": then,
					"modified_time": now,
				}
				if _, err := db.Put(context.Background(), "foo", doc); err != nil {
					t.Fatal(err)
				}
				return db
			}(),
			doc: &testDoc{ID: "foo", ITime: now, MTime: now},
			err: "document update conflict",
		},
		{
			name: "Merge fails",
			db: func() getPutter {
				db := testDB(t)
				doc := map[string]interface{}{
					"_id":           "foo",
					"imported_time": now,
					"modified_time": now,
				}
				if _, err := db.Put(context.Background(), "foo", doc); err != nil {
					t.Fatal(err)
				}
				return db
			}(),
			doc: &testDoc{ID: "foo", ITime: now, MTime: now, mergeErr: errors.New("merge error")},
			err: "failed to merge into existing document: merge error",
		},
		{
			name: "Merge no change",
			db: func() getPutter {
				db := testDB(t)
				doc := map[string]interface{}{
					"_id":           "foo",
					"imported_time": now,
					"modified_time": now,
				}
				if _, err := db.Put(context.Background(), "foo", doc); err != nil {
					t.Fatal(err)
				}
				return db
			}(),
			doc:      &testDoc{ID: "foo", Rev: "1", ITime: now, MTime: now},
			expected: map[string]interface{}{"_id": "foo", "_rev": "1", "imported_time": now, "modified_time": now},
		},
		{
			name: "Merge changed",
			db: func() getPutter {
				db := testDB(t)
				doc := map[string]interface{}{
					"_id":           "foo",
					"imported_time": then,
					"modified_time": then,
				}
				if _, err := db.Put(context.Background(), "foo", doc); err != nil {
					t.Fatal(err)
				}
				return db
			}(),
			doc:      &testDoc{ID: "foo", ITime: now, MTime: now, doMerge: true},
			expected: map[string]interface{}{"_id": "foo", "_rev": "2", "imported_time": now, "modified_time": now, "value": "new value"},
		},
		{
			name: "No change",
			db: func() kivikDB {
				db := testDB(t)
				doc := map[string]interface{}{
					"_id":           "foo",
					"imported_time": then,
					"modified_time": then,
				}
				if _, err := db.Put(context.Background(), "foo", doc); err != nil {
					t.Fatal(err)
				}
				return db
			}(),
			doc:      &testDoc{ID: "foo", ITime: then, MTime: then},
			expected: map[string]interface{}{"_id": "foo", "_rev": "1", "imported_time": then, "modified_time": then},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := saveDoc(context.Background(), test.db, test.doc)
			var msg string
			if err != nil {
				msg = err.Error()
			}
			if test.err != msg {
				t.Errorf("Unexpected error: %s", msg)
				return
			}
			if err != nil {
				return
			}
			row, err := test.db.Get(context.Background(), test.doc.DocID())
			if err != nil {
				t.Fatal(err)
			}
			var result map[string]interface{}
			if err := row.ScanDoc(&result); err != nil {
				t.Fatal(err)
			}
			revParts := strings.Split(result["_rev"].(string), "-")
			result["_rev"] = revParts[0]
			if d := diff.AsJSON(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

type mockGetter struct {
	row kivikRow
	err error
}

var _ getter = &mockGetter{}

func (db *mockGetter) Get(ctx context.Context, docID string, options ...kivik.Options) (kivikRow, error) {
	return db.row, db.err
}

type mockRow string

var _ kivikRow = mockRow("")

func (r mockRow) ScanDoc(i interface{}) error {
	return json.Unmarshal([]byte(r), &i)
}

func TestGetDoc(t *testing.T) {
	type gdTest struct {
		name     string
		db       getter
		id       string
		dst      interface{}
		expected interface{}
		err      string
	}
	tests := []gdTest{
		{
			name: "get fails",
			db:   &mockGetter{err: errors.New("get failed")},
			err:  "get failed",
		},
		{
			name: "Scan doc fails",
			db:   &mockGetter{row: mockRow("invalid JSON")},
			err:  "invalid character 'i' looking for beginning of value",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := getDoc(context.Background(), test.db, test.id, test.dst)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.AsJSON(test.expected, test.dst); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestFirstErr(t *testing.T) {
	type feTest struct {
		name     string
		errs     []error
		expected string
	}
	tests := []feTest{
		{
			name: "no errors",
		},
		{
			name: "only nil",
			errs: []error{nil},
		},
		{
			name:     "one err",
			errs:     []error{errors.New("one")},
			expected: "one",
		},
		{
			name:     "three errs",
			errs:     []error{errors.New("one"), errors.New("two"), errors.New("three")},
			expected: "one",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := firstErr(test.errs...)
			checkErr(t, test.expected, err)
		})
	}
}

func TestGetAttachment(t *testing.T) {
	tests := []struct {
		name     string
		db       attachmentGetter
		docID    string
		filename string
		expected *fb.Attachment
		err      string
	}{
		{
			name:     "not found",
			filename: "foo.txt",
			db:       &mockAttachmentGetter{err: errors.New("not found")},
			err:      "not found",
		},
		{
			name: "found it",
			db: &mockAttachmentGetter{attachments: map[string]*kivik.Attachment{
				"foo.txt": &kivik.Attachment{
					Filename:    "foo.txt",
					ContentType: "text/plain",
					ReadCloser:  ioutil.NopCloser(strings.NewReader("some text")),
				},
			}},
			filename: "foo.txt",
			expected: &fb.Attachment{
				ContentType: "text/plain",
				Content:     []byte("some text"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := getAttachment(context.Background(), test.db, test.docID, test.filename)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestUpdateDocs(t *testing.T) {
	tests := []struct {
		name string
		db   bulkDocer
		docs []interface{}
		err  string
	}{
		{
			name: "db error",
			db:   &mockBulkDocer{err: errors.New("db error")},
			docs: []interface{}{""},
			err:  "db error",
		},
		{
			name: "single doc error",
			db: &mockBulkDocer{
				results: &mockBulkResults{
					errs: []error{nil, errors.New("doc error"), nil},
				},
			},
			docs: []interface{}{"", "", ""},
			err:  "doc error",
		},
		{
			name: "result error",
			db: &mockBulkDocer{
				results: &mockBulkResults{
					errs: []error{nil, nil, nil},
					err:  errors.New("result error"),
				},
			},
			docs: []interface{}{"", "", ""},
			err:  "result error",
		},
		{
			name: "success",
			db: &mockBulkDocer{
				results: &mockBulkResults{
					errs: []error{nil, nil, nil},
				},
			},
			docs: []interface{}{"", "", ""},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := updateDocs(context.Background(), test.db, test.docs)
			checkErr(t, test.err, err)
		})
	}
}
