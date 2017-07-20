package model

import (
	"context"
	"encoding/json"
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
			checkDoc(t, udb, test.expected)
			checkDoc(t, bdb, test.expected)
		})
	}
}

func checkDoc(t *testing.T, db *kivik.DB, doc interface{}) {
	var docID string
	switch b := doc.(type) {
	case map[string]interface{}:
		docID = b["_id"].(string)
	case *fb.Bundle:
		docID = b.ID.String()
	default:
		x, err := json.Marshal(doc)
		if err != nil {
			panic(err)
		}
		var result struct {
			ID string `json:"_id"`
		}
		if e := json.Unmarshal(x, &result); e != nil {
			panic(e)
		}
		docID = result.ID
	}
	row, err := db.Get(context.Background(), docID)
	if err != nil {
		t.Errorf("failed to fetch %s: %s", docID, err)
		return
	}
	var result map[string]interface{}
	if err := row.ScanDoc(&result); err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(result["_rev"].(string), "-")
	result["_rev"] = parts[0]
	if d := diff.AsJSON(doc, result); d != "" {
		t.Error(d)
	}
}
