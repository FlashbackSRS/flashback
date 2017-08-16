package model

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
)

func ParseTime(ts string) time.Time {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		panic(err)
	}
	return t
}

type mockFile struct {
	body []byte
	err  error
}

var _ inputFile = &mockFile{}

func (f *mockFile) Bytes() ([]byte, error) { return f.body, f.err }

func TestImportFile(t *testing.T) {
	type ifTest struct {
		name string
		repo *Repo
		file inputFile
		err  string
	}
	tests := []ifTest{
		{
			name: "Not logged in",
			repo: &Repo{},
			err:  "not logged in",
		},
		{
			name: "file read error",
			repo: &Repo{user: "user-mjxwe"},
			file: &mockFile{err: errors.New("read error")},
			err:  "read error",
		},
		{
			name: "Invalid gzip data",
			repo: &Repo{user: "user-mjxwe"},
			file: &mockFile{body: []byte("bogus data")},
			err:  "gzip: invalid header",
		},
		{
			name: "Invalid Package JSON",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), "user-mjxwe"); err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), fb.EncodeDBID("bundle", []byte{1, 2, 3, 4})); err != nil {
					t.Fatal(err)
				}
				return &Repo{
					user:  "user-mjxwe",
					local: local,
				}
			}(),
			file: &mockFile{body: []byte{
				0x1f, 0x8b, 0x08, 0x08, 0xe2, 0x20, 0x71, 0x59,
				0x00, 0x03, 0x78, 0x00, 0x33, 0xe4, 0x02, 0x00,
				0x53, 0xfc, 0x51, 0x67, 0x02, 0x00, 0x00, 0x00}},
			err: "Unable to decode JSON: json: cannot unmarshal number into Go value of type fb.jsonPackage",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var msg string
			if err := test.repo.ImportFile(context.Background(), test.file); err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s", msg)
			}
			if msg != "" {
				return
			}
		})
	}
}

// func MustParseDbID(id string) fb.DbID {
// 	dbid, err := fb.ParseDbID(id)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return dbid
// }

func TestImport(t *testing.T) {
	type iTest struct {
		name           string
		repo           *Repo
		file           io.Reader
		expectedBundle *fb.Bundle
		expectedDocs   []interface{}
		err            string
	}
	tests := []iTest{
		{
			name: "Not logged in",
			repo: &Repo{},
			file: strings.NewReader("{}"),
			err:  "not logged in",
		},
		{
			name: "Invalid JSON",
			repo: &Repo{user: "user-mjxwe",
				local: func() kivikClient {
					c := testClient(t)
					if err := c.CreateDB(context.Background(), "user-mjxwe"); err != nil {
						t.Fatal(err)
					}
					return c
				}(),
			},
			file: strings.NewReader("bogus data"),
			err:  "Unable to decode JSON: invalid character 'b' looking for beginning of value",
		},
		{
			name: "Invalid package",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), "user-mjxwe"); err != nil {
					t.Fatal(err)
				}
				return &Repo{
					user:  "user-mjxwe",
					local: local,
				}
			}(),
			file: strings.NewReader(`{"version":0}`),
			err:  "Unable to decode JSON: package version 0 < 1",
		},
		{
			name: "Valid",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), "user-mjxwe"); err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), "bundle-aebagba"); err != nil {
					t.Fatal(err)
				}
				return &Repo{
					user:  "user-mjxwe",
					local: local,
				}
			}(),
			file: func() io.Reader {
				return strings.NewReader(`
				{
					"version": 2,
					"bundle": {
						"_id": "bundle-aebagba",
						"type": "bundle",
						"created": "2016-07-31T15:08:24.730156517Z",
						"modified": "2016-07-31T15:08:24.730156517Z",
						"owner": "user-mjxwe"
					},
					"cards": [
						{
							"type": "card",
							"_id": "card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.0",
							"created": "2016-07-31T15:08:24.730156517Z",
							"modified": "2016-07-31T15:08:24.730156517Z",
							"model": "theme-VGVzdCBUaGVtZQ/0"
						}
					],
					"notes": [
						{
							"_id": "note-VGVzdCBOb3Rl",
							"type": "note",
							"created": "2016-07-31T15:08:24.730156517Z",
							"modified": "2016-07-31T15:08:24.730156517Z",
							"imported": "2016-08-02T15:08:24.730156517Z",
							"theme": "theme-VGVzdCBUaGVtZQ",
							"model": 0,
							"fieldValues": [
								{
									"text": "cat"
								}
							]
						}
					],
					"decks": [
						{
							"_id": "deck-VGVzdCBEZWNr",
							"type": "deck",
							"created": "2016-07-31T15:08:24.730156517Z",
							"modified": "2016-07-31T15:08:24.730156517Z",
							"imported": "2016-08-02T15:08:24.730156517Z",
							"name": "Test Deck",
							"description": "Deck for testing",
							"cards": ["card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.0"]
						}
					],
					"themes": [
						{
							"_id": "theme-VGVzdCBUaGVtZQ",
							"type": "theme",
							"created": "2016-07-31T15:08:24.730156517Z",
							"modified": "2016-07-31T15:08:24.730156517Z",
							"imported": "2016-08-02T15:08:24.730156517Z",
							"name": "Test Theme",
							"description": "Theme for testing",
							"models": [
								{
									"id": 0,
									"modelType": "anki-basic",
									"name": "Model A",
									"templates": [],
									"fields": [
										{
											"fieldType": 0,
											"name": "Word"
										}
									],
									"files": [
										"m1.html"
									]
								}
							],
							"_attachments": {
								"$main.css": {
									"content_type": "text/css",
									"data": "LyogYW4gZW1wdHkgQ1NTIGZpbGUgKi8="
								},
								"m1.html": {
									"content_type": "text/html",
									"data": "PGh0bWw+PC9odG1sPg=="
								}
							},
							"files": [
								"$main.css"
							],
							"modelSequence": 2
						}
					],
					"reviews": [
						{
							"cardID": "card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.0",
							"timestamp": "2017-01-01T01:01:01Z"
						}
					]
				}
				`)
			}(),
			expectedBundle: &fb.Bundle{
				ID:       "bundle-aebagba",
				Rev:      "1",
				Owner:    "user-mjxwe",
				Created:  ParseTime("2016-07-31T15:08:24.730156517Z"),
				Modified: ParseTime("2016-07-31T15:08:24.730156517Z"),
			},
			expectedDocs: []interface{}{
				map[string]interface{}{
					"_id":      "card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.0",
					"_rev":     "1",
					"type":     "card",
					"created":  "2016-07-31T15:08:24.730156517Z",
					"modified": "2016-07-31T15:08:24.730156517Z",
					"model":    "theme-VGVzdCBUaGVtZQ/0",
				},
				map[string]interface{}{
					"_id":         "deck-VGVzdCBEZWNr",
					"type":        "deck",
					"_rev":        "1",
					"name":        "Test Deck",
					"description": "Deck for testing",
					"cards":       []string{"card-krsxg5baij2w4zdmmu.VGVzdCBOb3Rl.0"},
					"created":     "2016-07-31T15:08:24.730156517Z",
					"imported":    "2016-08-02T15:08:24.730156517Z",
					"modified":    "2016-07-31T15:08:24.730156517Z",
				},
				map[string]interface{}{
					"_id":           "theme-VGVzdCBUaGVtZQ",
					"type":          "theme",
					"_rev":          "1",
					"name":          "Test Theme",
					"description":   "Theme for testing",
					"modelSequence": 2,
					"created":       "2016-07-31T15:08:24.730156517Z",
					"imported":      "2016-08-02T15:08:24.730156517Z",
					"modified":      "2016-07-31T15:08:24.730156517Z",
					"models": []map[string]interface{}{
						{
							"id":        0,
							"modelType": "anki-basic",
							"name":      "Model A",
							"templates": []string{},
							"fields": []map[string]interface{}{
								{
									"fieldType": 0,
									"name":      "Word",
								},
							},
							"files": []string{"m1.html"},
						},
					},
					"files": []string{"$main.css"},
				},
				map[string]interface{}{
					"_id":         "note-VGVzdCBOb3Rl",
					"type":        "note",
					"_rev":        "1",
					"theme":       "theme-VGVzdCBUaGVtZQ",
					"model":       0,
					"fieldValues": []map[string]string{{"text": "cat"}},
					"created":     "2016-07-31T15:08:24.730156517Z",
					"imported":    "2016-08-02T15:08:24.730156517Z",
					"modified":    "2016-07-31T15:08:24.730156517Z",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var msg string
			err := test.repo.Import(context.Background(), test.file)
			if err != nil {
				msg = err.Error()
			}
			if test.err != msg {
				t.Errorf("Unexpected error: %s", msg)
			}
			if err != nil {
				return
			}
			udb, err := test.repo.userDB(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			checkDoc(t, udb, test.expectedBundle)
			bdb, err := test.repo.bundleDB(context.Background(), test.expectedBundle)
			if err != nil {
				t.Fatal(err)
			}
			checkDoc(t, bdb, test.expectedBundle)
			for _, doc := range test.expectedDocs {
				checkDoc(t, udb, doc)
			}
		})
	}
}

/*
var UUID = []byte{0xD1, 0xC9, 0x58, 0x7D, 0x88, 0xDF, 0x4A, 0x65, 0x89, 0x23, 0xF7, 0x3C, 0xDF, 0x6D, 0x1D, 0x70}

func BenchmarkSaveCard(b *testing.B) {
	u, err := fb.NewUser(uuid.UUID(UUID), "testuser")
	if err != nil {
		panic(err)
	}
	user := &User{u}
	client, _ := kivik.New(context.TODO(), "pouch", "")
	if err = client.DestroyDB(context.TODO(), u.ID.String()); err != nil {
		panic(err)
	}
	db, err := user.DB()
	if err != nil {
		panic(err)
	}
	err = db.CreateIndex(context.TODO(), "", "", map[string]interface{}{
		"fields": []string{"due", "created", "type"},
	})
	if err != nil {
		panic(err)
	}
	cards := make([]*fb.Card, b.N)
	for i := 0; i < b.N; i++ {
		id := fmt.Sprintf("card-bundle.%x.0", i)
		card, _ := fb.NewCard("themefoo", 0, id)
		cards[i] = card
	}
	b.ResetTimer()
	for _, card := range cards {
		if err := db.Save(card); err != nil {
			panic(err)
		}
	}
}
*/

type failBulkDocs struct {
	kivik.DB
}

func (f *failBulkDocs) BulkDocs(_ context.Context, _ interface{}) (*kivik.BulkResults, error) {
	return nil, errors.New("bulkdocs failed")
}

func (f *failBulkDocs) Get(_ context.Context, _ string, _ ...kivik.Options) (kivikRow, error) {
	return nil, nil
}

func TestBulkInsert(t *testing.T) {
	type biTest struct {
		name     string
		db       getPutBulkDocer
		docs     []FlashbackDoc
		expected []map[string]interface{}
		err      string
	}
	now := time.Now()
	tests := []biTest{
		{
			name: "BulkDocs fails",
			db:   &failBulkDocs{},
			err:  "bulkdocs failed",
		},
		{
			name: "no documents",
		},
		{
			name: "new docs",
			docs: []FlashbackDoc{
				&testDoc{ID: "abc", Value: "foo"},
				&testDoc{ID: "def", Value: "bar"},
			},
			expected: []map[string]interface{}{
				{"_id": "abc", "_rev": "1", "value": "foo", "imported_time": time.Time{}, "modified_time": time.Time{}},
				{"_id": "def", "_rev": "1", "value": "bar", "imported_time": time.Time{}, "modified_time": time.Time{}},
			},
		},
		{
			name: "new and conflict",
			db: func() getPutBulkDocer {
				db := testDB(t)
				if _, err := db.Put(context.Background(), "abc", map[string]interface{}{"_id": "abc", "value": "foo"}); err != nil {
					t.Fatal(err)
				}
				return db
			}(),
			docs: []FlashbackDoc{
				&testDoc{ID: "abc", Value: "foo"},
				&testDoc{ID: "def", Value: "bar"},
			},
			err: "1 error occurred:\n\n* failed to save doc abc: document update conflict",
			expected: []map[string]interface{}{
				{"_id": "abc", "_rev": "1", "value": "foo"},
				{"_id": "def", "_rev": "1", "value": "bar", "imported_time": time.Time{}, "modified_time": time.Time{}},
			},
		},
		{
			name: "new and merge",
			db: func() getPutBulkDocer {
				db := testDB(t)
				ctime := time.Now().Add(-time.Hour)
				doc := testDoc{
					ID:    "abc",
					ITime: ctime,
					MTime: ctime,
					Value: "foo",
				}
				if _, err := db.Put(context.Background(), "abc", doc); err != nil {
					t.Fatal(err)
				}
				return db
			}(),
			docs: []FlashbackDoc{
				&testDoc{ID: "abc", Value: "foo", ITime: now, doMerge: true},
				&testDoc{ID: "def", Value: "bar", ITime: now},
			},
			expected: []map[string]interface{}{
				{"_id": "abc", "_rev": "2", "value": "new value", "imported_time": now, "modified_time": time.Time{}},
				{"_id": "def", "_rev": "1", "value": "bar", "imported_time": now, "modified_time": time.Time{}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := test.db
			if db == nil {
				db = testDB(t)
			}
			var msg string
			if err := bulkInsert(context.Background(), db, test.docs...); err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s", msg)
			}
			for i, expected := range test.expected {
				row, err := db.Get(context.Background(), expected["_id"].(string))
				if err != nil {
					t.Fatal(err)
				}
				var result map[string]interface{}
				if e := row.ScanDoc(&result); e != nil {
					t.Fatal(e)
				}
				revParts := strings.Split(result["_rev"].(string), "-")
				result["_rev"] = revParts[0]
				if d := diff.AsJSON(expected, result); d != nil {
					t.Errorf("Doc %d: %s", i, d)
				}
			}
		})
	}
}
