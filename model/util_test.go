package model

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

type testDoc struct {
	ID       string     `json:"_id"`
	Rev      string     `json:"_rev,omitempty"`
	ITime    *time.Time `json:"imported_time,omitempty"`
	MTime    *time.Time `json:"modified_time,omitempty"`
	Value    string     `json:"value,omitempty"`
	doMerge  bool
	mergeErr error
}

var _ FlashbackDoc = &testDoc{}

func (d *testDoc) DocID() string                { return d.ID }
func (d *testDoc) SetRev(rev string)            { d.Rev = rev }
func (d *testDoc) ImportedTime() *time.Time     { return d.ITime }
func (d *testDoc) ModifiedTime() *time.Time     { return d.MTime }
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

var _ saveDB = &mockDB{}

func (db *mockDB) Get(_ context.Context, _ string, _ ...kivik.Options) (*kivik.Row, error) {
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
		db       saveDB
		doc      FlashbackDoc
		expected map[string]interface{}
		err      string
	}
	now := time.Now()
	then := time.Now().Add(-24 * time.Hour)
	tests := []sdTest{
		{
			name:     "New doc",
			db:       testDB(t),
			doc:      &testDoc{ID: "foo"},
			expected: map[string]interface{}{"_id": "foo", "_rev": "1"},
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
			doc: &testDoc{ID: "_foo", ITime: &now},
			err: "failed to fetch existing document: get error",
		},
		{
			name: "Conflict with non-imported doc",
			db: func() *kivik.DB {
				db := testDB(t)
				doc := map[string]string{
					"_id": "foo",
				}
				if _, err := db.Put(context.Background(), "foo", doc); err != nil {
					t.Fatal(err)
				}
				return db
			}(),
			doc: &testDoc{ID: "foo", ITime: &now},
			err: "document update conflict",
		},
		{
			name: "Modified Doc Exists",
			db: func() *kivik.DB {
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
			doc: &testDoc{ID: "foo", ITime: &now, MTime: &now},
			err: "document update conflict",
		},
		{
			name: "Merge fails",
			db: func() *kivik.DB {
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
			doc: &testDoc{ID: "foo", ITime: &now, MTime: &now, mergeErr: errors.New("merge error")},
			err: "failed to merge into existing document: merge error",
		},
		{
			name: "Merge no change",
			db: func() *kivik.DB {
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
			doc:      &testDoc{ID: "foo", Rev: "1", ITime: &now, MTime: &now},
			expected: map[string]interface{}{"_id": "foo", "_rev": "1", "imported_time": now, "modified_time": now},
		},
		{
			name: "Merge changed",
			db: func() *kivik.DB {
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
			doc:      &testDoc{ID: "foo", ITime: &now, MTime: &now, doMerge: true},
			expected: map[string]interface{}{"_id": "foo", "_rev": "2", "imported_time": now, "modified_time": now, "value": "new value"},
		},
		{
			name: "No change",
			db: func() *kivik.DB {
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
			doc:      &testDoc{ID: "foo", ITime: &then, MTime: &then},
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
			if d := diff.AsJSON(test.expected, result); d != "" {
				t.Error(d)
			}
		})
	}
}
