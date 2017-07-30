package model

import (
	"context"
	"errors"
	"testing"

	"github.com/flimzy/kivik"
)

func TestDbDSN(t *testing.T) {
	db := testDB(t)
	result := dbDSN(db)
	expected := "local/" + db.Name()
	if result != expected {
		t.Errorf("Unexpected result: %s\n", result)
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
			err: "sync local to remote: kivik: driver does not support replication",
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
