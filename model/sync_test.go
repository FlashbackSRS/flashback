package model

import (
	"context"
	"errors"
	"testing"
	"time"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/testy"
)

func TestDbDSN(t *testing.T) {
	db := testDB(t)
	result := dbDSN(db)
	expected := db.Name()
	if result != expected {
		t.Errorf("Expected: %s\n  Actual: %s\n", expected, result)
	}
}

func TestSync(t *testing.T) {
	type sTest struct {
		name string
		repo *Repo
		err  string
	}
	tests := []sTest{
		{
			name: "not logged in",
			repo: &Repo{},
			err:  "not logged in",
		},
		// {
		// 	name: "logged in",
		// 	repo: func() *Repo {
		// 		local, err := localConnection()
		// 		if err != nil {
		// 			t.Fatal(err)
		// 		}
		// 		if e := local.CreateDB(context.Background(), "user-bob"); e != nil {
		// 			t.Fatal(e)
		// 		}
		// 		remote, err := remoteConnection("")
		// 		if err != nil {
		// 			t.Fatal(err)
		// 		}
		// 		if e := remote.CreateDB(context.Background(), "user-bob"); e != nil {
		// 			t.Fatal(e)
		// 		}
		// 		return &Repo{
		// 			user:   "bob",
		// 			local:  local,
		// 			remote: remote,
		// 		}
		// 	}(),
		// 	err: func() string {
		// 		if env == "js" {
		// 			return ""
		// 		}
		// 		return "sync local to remote: kivik: driver does not support replication"
		// 	}(),
		// },
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var msg string
			if err := test.repo.Sync(context.Background()); err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s", msg)
			}
		})
	}
}

type fakeReplicator struct {
	*kivik.Client
	err error
}

func (r *fakeReplicator) Replicate(_ context.Context, _, _ string, _ ...kivik.Options) (*kivik.Replication, error) {
	return nil, r.err
}

func TestReplicate(t *testing.T) {
	type rTest struct {
		name           string
		client         clientReplicator
		target, source string
		err            string
	}
	tests := []rTest{
		{
			name:   "Replication fails",
			client: &fakeReplicator{err: errors.New("replication failed")},
			err:    "replication failed",
		},
		{
			name:   "Replication succeeds",
			client: &fakeReplicator{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var count int32
			var msg string
			if err := replicate(context.Background(), test.client, test.target, test.source, &count); err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected failure: %s", msg)
			}
		})
	}
}

type fakeReplication struct {
	updates        int
	err, updateErr error
	docsWritten    int64
}

var _ replication = &fakeReplication{}

func (r *fakeReplication) Delete(_ context.Context) error { r.updates = 0; return nil }
func (r *fakeReplication) IsActive() bool                 { return r.updates > 0 }
func (r *fakeReplication) Err() error                     { return r.err }
func (r *fakeReplication) DocsWritten() int64             { return r.docsWritten }
func (r *fakeReplication) Update(_ context.Context) error { r.updates--; return r.updateErr }

func TestProcessReplication(t *testing.T) {
	type prTest struct {
		name  string
		rep   replication
		count int32
		err   string
	}
	tests := []prTest{
		{
			name: "replication failure",
			rep:  &fakeReplication{err: errors.New("replication failure")},
			err:  "replication failure",
		},
		{
			name: "update failure",
			rep:  &fakeReplication{updates: 1, updateErr: errors.New("update failure")},
			err:  "update failure",
		},
		{
			name:  "success",
			count: 10,
			rep:   &fakeReplication{updates: 1, docsWritten: 10},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			count, err := processReplication(context.Background(), test.rep)
			var msg string
			if err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s", msg)
			}
			if count != test.count {
				t.Errorf("Unexpected result: %d", count)
			}
		})
	}
}

func TestSyncBundles(t *testing.T) {
	type sbTest struct {
		name          string
		repo          *Repo
		reads, writes int32
		err           string
	}
	tests := []sbTest{
		{
			name: "not logged in",
			repo: &Repo{},
			err:  "not logged in",
		},
		// TODO: Test actual replications
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()
			var reads, writes int32
			err := test.repo.syncBundles(ctx, &reads, &writes)
			var msg string
			if err != nil {
				msg = err.Error()
			}
			if test.err != msg {
				t.Errorf("Unexpected error: %s", msg)
			}
			if err != nil {
				return
			}
			if test.reads != reads || test.writes != writes {
				t.Errorf("Unexpected sync count.\nExpected reads %d, writes %d\n  Actual reads %d, writes %d", test.reads, test.writes, reads, writes)
			}
		})
	}
}

func TestLastSyncTime(t *testing.T) {
	type lstTest struct {
		name         string
		repo         *Repo
		status       int
		expectedRev  string
		expectedTime time.Time
	}
	tests := []lstTest{
		{
			name:   "not logged in",
			repo:   &Repo{},
			status: kivik.StatusUnauthorized,
		},
		{
			name: "db does not exist",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				return &Repo{
					user:  "bob0",
					local: local,
				}
			}(),
			status: kivik.StatusNotFound,
		},
		{
			name: "doc not found",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if e := local.CreateDB(context.Background(), "user-bob1"); e != nil {
					t.Fatal(e)
				}
				return &Repo{
					user:  "bob1",
					local: local,
				}
			}(),
			status: kivik.StatusNotFound,
		},
		{
			name: "invalid JSON response",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if e := local.CreateDB(context.Background(), "user-bob2"); e != nil {
					t.Fatal(e)
				}
				db, err := local.DB(context.Background(), "user-bob2")
				if err != nil {
					t.Fatal(err)
				}
				doc := map[string]string{"lastSync": "foo"}
				if _, e := db.Put(context.Background(), lastSyncTimestampDocID, doc); e != nil {
					t.Fatal(e)
				}
				return &Repo{
					user:  "bob2",
					local: local,
				}
			}(),
			status: kivik.StatusInternalServerError,
		},
		func() lstTest {
			local, err := localConnection()
			if err != nil {
				t.Fatal(err)
			}
			if e := local.CreateDB(context.Background(), "user-bob3"); e != nil {
				t.Fatal(e)
			}
			db, err := local.DB(context.Background(), "user-bob3")
			if err != nil {
				t.Fatal(err)
			}
			ts, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z07:00")
			doc := map[string]interface{}{"lastSync": ts}
			rev, err := db.Put(context.Background(), lastSyncTimestampDocID, doc)
			if err != nil {
				t.Fatal(err)
			}
			return lstTest{
				name:         "success",
				expectedRev:  rev,
				expectedTime: ts,
				repo: &Repo{
					user:  "bob3",
					local: local,
				},
			}
		}(),
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rev, result, err := test.repo.lastSyncTime(context.Background())
			var status int
			if err != nil {
				status = kivik.StatusCode(err)
			}
			if status != test.status {
				t.Errorf("Unexpected error: %d %s", status, err)
			}
			if err != nil {
				return
			}
			if test.expectedRev != rev {
				t.Errorf("Unexpected rev: %s (want %s)", rev, test.expectedRev)
			}
			if !test.expectedTime.Equal(result) {
				t.Errorf("Unexpected result: %v", result)
			}
		})
	}
}

func TestUpdateSyncTime(t *testing.T) {
	tests := []struct {
		name         string
		repo         *Repo
		dbname       string
		err          string
		expectedTime time.Time
	}{
		{
			name: "not logged in",
			repo: &Repo{},
			err:  "not logged in",
		},
		{
			name: "first sync",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if e := local.CreateDB(context.Background(), "user-bob"); e != nil {
					t.Fatal(e)
				}
				return &Repo{
					user:  "bob",
					local: local,
				}
			}(),
			expectedTime: func() time.Time {
				t, _ := time.Parse(time.RFC3339, "2017-01-01T12:00:00Z")
				return t
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var errMsg string
			err := test.repo.updateSyncTime(context.Background())
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != test.err {
				t.Errorf("Unexpected error: %s", errMsg)
			}
			if err != nil {
				return
			}
			_, ts, err := test.repo.lastSyncTime(context.Background())
			if !ts.Equal(test.expectedTime) {
				t.Errorf("Unexpected time stored: %v", ts)
			}
		})
	}
}

func TestUpgradeSchema(t *testing.T) {
	tests := []struct {
		name     string
		repo     *Repo
		expected bool
		err      string
		check    func(*Repo) error
	}{
		{
			name: "not logged in",
			repo: &Repo{},
			err:  "failed to connect to db: not logged in",
		},
		// {
		// 	name: "card to update",
		// 	repo: func() *Repo {
		// 		local, err := localConnection()
		// 		if err != nil {
		// 			t.Fatal(err)
		// 		}
		// 		if e := local.CreateDB(context.Background(), "user-bob"); e != nil {
		// 			t.Fatal(e)
		// 		}
		// 		db, err := local.DB(context.Background(), "user-bob")
		// 		if err != nil {
		// 			t.Fatal(err)
		// 		}
		// 		card := map[string]string{"_id": "card-mzxw6cq.bmlsCg.0"}
		// 		if _, e := db.Put(context.Background(), card["_id"], card); e != nil {
		// 			t.Fatal(e)
		// 		}
		// 		deck := map[string]interface{}{
		// 			"_id":   "deck-1234",
		// 			"cards": []string{"card-mzxw6cq.bmlsCg.0"},
		// 		}
		// 		if e := local.CreateDB(context.Background(), "bundle-mzxw6cq"); e != nil {
		// 			t.Fatal(e)
		// 		}
		// 		bdb, err := local.DB(context.Background(), "bundle-mzxw6cq")
		// 		if err != nil {
		// 			t.Fatal(err)
		// 		}
		// 		if _, e := bdb.Put(context.Background(), deck["_id"].(string), deck); e != nil {
		// 			t.Fatal(e)
		// 		}
		// 		return &Repo{
		// 			user:  "bob",
		// 			local: local,
		// 		}
		// 	}(),
		// 	expected: true,
		// },
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.repo.upgradeSchema(context.Background())
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != test.err {
				t.Errorf("Unexpected error: %s", errMsg)
			}
			if err != nil {
				return
			}
			if test.expected != result {
				t.Errorf("Unexpected result: %t", result)
			}
			if test.check != nil {
				if e := test.check(test.repo); e != nil {
					t.Errorf("Check failed: %s", e)
				}
			}
		})
	}
}

func TestReadBundle(t *testing.T) {
	tests := []struct {
		name     string
		cache    *cardDeckCache
		bundleID string
		expected *cardDeckCache
		err      string
	}{
		{
			name: "db not found",
			cache: &cardDeckCache{
				client: testClient(t),
			},
			bundleID: "bundle-0000",
			err:      "database does not exist",
		},
		{
			name: "query error",
			cache: newCardDeckCache(&mockClient{
				db: &mockDB{alldocsErr: errors.New("alldocs error")},
			}),
			err: "alldocs error",
		},
		{
			name: "bad deck JSON",
			cache: newCardDeckCache(&mockClient{
				db: &mockDB{
					alldocsRows: &mockRows{
						rows: []string{"invalid json"},
					},
				},
			}),
			err: "invalid character 'i' looking for beginning of value",
		},
		{
			name:     "success",
			bundleID: "bundle-0000",
			cache: newCardDeckCache(&mockClient{
				db: &mockDB{
					alldocsRows: &mockRows{
						rows: []string{`{
								"_id":         "deck-ZGVjaw",
								"type":        "deck",
								"name":        "test name",
								"description": "test description",
								"created":     "2017-01-01T00:00:00Z",
								"modified":    "2017-01-01T00:00:00Z",
								"imported":    "2017-01-01T00:00:00Z",
								"cards":       ["card-YmFy.bmlsCg.0","card-Zm9v.bmlsCg.0"]
							}`,
						},
					},
				},
			}),
			expected: &cardDeckCache{
				cache: map[string]string{
					"card-YmFy.bmlsCg.0": "deck-ZGVjaw",
					"card-Zm9v.bmlsCg.0": "deck-ZGVjaw",
				},
				readBundles: map[string]struct{}{"bundle-0000": struct{}{}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.cache.readBundle(context.Background(), test.bundleID)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != test.err {
				t.Errorf("Unexpected error: %s", errMsg)
			}
			if err != nil {
				return
			}
			test.cache.client = nil // Don't care about this for the check
			if d := diff.Interface(test.expected, test.cache); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestCacheCardDeck(t *testing.T) {
	tests := []struct {
		name     string
		card     *fb.Card
		cache    *cardDeckCache
		expected string
		err      string
	}{
		{
			name: "readBundle error",
			card: &fb.Card{
				ID: "card-YmFy.bmlsCg.0",
			},
			cache: &cardDeckCache{
				client: testClient(t),
			},
			err: "database does not exist",
		},
		{
			name: "cached value",
			card: &fb.Card{
				ID: "card-YmFy.bmlsCg.0",
			},
			cache: &cardDeckCache{
				cache: map[string]string{
					"card-YmFy.bmlsCg.0": "deck-ZGVjaw",
					"card-Zm9v.bmlsCg.0": "deck-ZGVjaw",
				},
				readBundles: map[string]struct{}{"bundle-YmFy": struct{}{}},
			},
			expected: "deck-ZGVjaw",
		},
		{
			name: "card without deck",
			card: &fb.Card{
				ID: "card-YmFy.bmlsCg.0",
			},
			cache: &cardDeckCache{
				cache:       map[string]string{},
				readBundles: map[string]struct{}{"bundle-YmFy": struct{}{}},
			},
			expected: orphanedCardDeck,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.cache.cardDeck(context.Background(), test.card)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != test.err {
				t.Errorf("Unexpected error: %s", errMsg)
			}
			if err != nil {
				return
			}
			if result != test.expected {
				t.Errorf("Unexpected result: %s", result)
			}
		})
	}
}

func TestStoreDecksInUserDB(t *testing.T) {
	tests := []struct {
		name     string
		repo     *Repo
		expected bool
		err      string
		verify   func(*Repo) error
	}{
		{
			name: "not logged in",
			repo: &Repo{},
			err:  "user db: not logged in",
		},
		// {
		// 	name: "no bundles",
		// 	repo: &Repo{
		// 		user:  "bob",
		// 		local: &mockClient{db: &mockAllDocer{}},
		// 	},
		// 	expected: false,
		// },
		{
			name: "missing bundle",
			repo: &Repo{
				user: "bob",
				local: &mockClient{
					dbs: map[string]kivikDB{
						"user-bob": &mockAllDocer{
							rows: &mockRows{
								rows: []string{""},
								keys: []string{"bundle-foo"},
							},
						},
					},
				},
			},
			err: "mock db not found",
		},
		{
			name: "Alldocs failure",
			repo: &Repo{
				user: "bob",
				local: &mockClient{
					db: &mockAllDocer{
						err: errors.New("alldocs failed"),
					},
				},
			},
			err: "alldocs failed",
		},
		{
			name: "rows failure",
			repo: &Repo{
				user: "bob",
				local: &mockClient{
					db: &mockAllDocer{
						rows: &mockRows{
							err: errors.New("rows error"),
						},
					},
				},
			},
			err: "rows error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := storeDecksInUserDB(context.Background(), test.repo)
			testy.Error(t, test.err, err)
			if result != test.expected {
				t.Errorf("Unexpected result: %t", result)
			}
			if test.verify != nil {
				if err := test.verify(test.repo); err != nil {
					t.Errorf("Verification failed: %s", err)
				}
			}
		})
	}
}

func TestGetBundleIDs(t *testing.T) {
	tests := []struct {
		name     string
		db       kivikDB
		expected []string
		err      string
	}{
		{
			name: "Alldocs failure",
			db: &mockAllDocer{
				err: errors.New("alldocs failed"),
			},
			err: "alldocs failed",
		},
		{
			name: "rows failure",
			db: &mockAllDocer{
				rows: &mockRows{
					err: errors.New("rows error"),
				},
			},
			err: "rows error",
		},
		{
			name: "two bundles",
			db: &mockAllDocer{
				rows: &mockRows{
					rows: []string{"", ""},
					keys: []string{"bundle-foo", "bundle-bar"},
				},
			},
			expected: []string{"bundle-foo", "bundle-bar"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := getBundleIDs(context.Background(), test.db)
			testy.Error(t, test.err, err)
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
