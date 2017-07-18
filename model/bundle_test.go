package model

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
)

func TestSaveBundle(t *testing.T) {
	type sbTest struct {
		name     string
		repo     *Repo
		bundle   *fb.Bundle
		expected map[string]interface{}
		err      string
	}
	id, _ := fb.NewDbID("bundle", []byte{1, 2, 3, 4})
	owner, _ := fb.NewDbID("user", []byte{1, 2, 3, 4, 5})
	tests := []sbTest{
		{
			name:   "not logged in",
			repo:   &Repo{},
			bundle: &fb.Bundle{ID: id},
			err:    "not logged in",
		},
		{
			name:   "invalid bundle",
			repo:   &Repo{},
			bundle: &fb.Bundle{},
			err:    "invalid bundle",
		},
		{
			name: "user db does not exist",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				return &Repo{
					local: local,
					user:  "bob",
				}
			}(),
			bundle: &fb.Bundle{ID: id},
			err:    "userDB: database does not exist",
		},
		{
			name: "bundle db does not exist",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), "bob"); err != nil {
					t.Fatal(err)
				}
				return &Repo{
					local: local,
					user:  "bob",
				}
			}(),
			bundle: &fb.Bundle{ID: id},
			err:    "bundleDB: database does not exist",
		},
		{
			name: "success",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), "bob"); err != nil {
					t.Fatal(err)
				}
				if err := local.CreateDB(context.Background(), id.String()); err != nil {
					t.Fatal(err)
				}
				return &Repo{
					local: local,
					user:  "bob",
				}
			}(),
			bundle: &fb.Bundle{ID: id, Owner: &fb.User{ID: owner}},
			expected: map[string]interface{}{
				"_id":      id.String(),
				"_rev":     "1",
				"type":     "bundle",
				"owner":    owner.Identity(),
				"created":  time.Time{},
				"modified": time.Time{},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.repo.SaveBundle(context.Background(), test.bundle)
			var msg string
			if err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s", msg)
				return
			}
			if err != nil {
				return
			}
			udb, err := test.repo.userDB(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			bdb, err := test.repo.bundleDB(context.Background(), test.bundle)
			if err != nil {
				t.Fatal(err)
			}
			checkBundle(t, udb, test.expected)
			checkBundle(t, bdb, test.expected)
		})
	}
}

func checkBundle(t *testing.T, db *kivik.DB, bundle interface{}) {
	var bundleID string
	switch b := bundle.(type) {
	case map[string]interface{}:
		bundleID = b["_id"].(string)
	case *fb.Bundle:
		bundleID = b.ID.String()
	default:
		panic(fmt.Sprintf("Unknown type: %t", bundle))
	}
	row, err := db.Get(context.Background(), bundleID)
	if err != nil {
		t.Fatal(err)
	}
	var result map[string]interface{}
	if err := row.ScanDoc(&result); err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(result["_rev"].(string), "-")
	result["_rev"] = parts[0]
	if d := diff.AsJSON(bundle, result); d != "" {
		t.Error(d)
	}
}
