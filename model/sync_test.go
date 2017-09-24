package model

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/flimzy/kivik"
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
		{
			name: "logged in",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if e := local.CreateDB(context.Background(), "user-bob"); e != nil {
					t.Fatal(e)
				}
				remote, err := remoteConnection("")
				if err != nil {
					t.Fatal(err)
				}
				if e := remote.CreateDB(context.Background(), "user-bob"); e != nil {
					t.Fatal(e)
				}
				return &Repo{
					user:   "bob",
					local:  local,
					remote: remote,
				}
			}(),
			err: func() string {
				if env == "js" {
					return ""
				}
				return "sync local to remote: kivik: driver does not support replication"
			}(),
		},
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
		err          string
		expectedRev  string
		expectedTime time.Time
	}
	tests := []lstTest{
		{
			name: "not logged in",
			repo: &Repo{},
			err:  "not logged in",
		},
		{
			name: "db does not exist",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				return &Repo{
					user:  "bob",
					local: local,
				}
			}(),
			err: "database does not exist",
		},
		{
			name: "doc not found",
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
			err: "missing",
		},
		{
			name: "invalid JSON response",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if e := local.CreateDB(context.Background(), "user-bob"); e != nil {
					t.Fatal(e)
				}
				db, err := local.DB(context.Background(), "user-bob")
				if err != nil {
					t.Fatal(err)
				}
				doc := map[string]string{"lastSync": "foo"}
				if _, e := db.Put(context.Background(), lastSyncTimestampDocID, doc); e != nil {
					t.Fatal(e)
				}
				return &Repo{
					user:  "bob",
					local: local,
				}
			}(),
			err: `parsing time ""foo"" as ""2006-01-02T15:04:05Z07:00"": cannot parse "foo"" as "2006"`,
		},
		func() lstTest {
			local, err := localConnection()
			if err != nil {
				t.Fatal(err)
			}
			if e := local.CreateDB(context.Background(), "user-bob"); e != nil {
				t.Fatal(e)
			}
			db, err := local.DB(context.Background(), "user-bob")
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
					user:  "bob",
					local: local,
				},
			}
		}(),
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rev, result, err := test.repo.lastSyncTime(context.Background())
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
			if test.expectedRev != rev {
				t.Errorf("Unexpected rev: %s", rev)
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
