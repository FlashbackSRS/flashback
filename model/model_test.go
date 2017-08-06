package model

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"

	fb "github.com/FlashbackSRS/flashback-model"
)

func TestNew(t *testing.T) {
	t.Run("InvalidURL", func(t *testing.T) {
		_, err := New(context.Background(), "http://foo.com/%xx")
		if err == nil || err.Error() != `parse http://foo.com/%xx: invalid URL escape "%xx"` {
			t.Errorf("Unexpected error: %s", err)
		}
	})
	t.Run("Valid", func(t *testing.T) {
		_, err := New(context.Background(), "http://foo.com")
		if err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
	})
}

func TestAuth(t *testing.T) {
	s := mockServer(t)
	t.Run("Success", func(t *testing.T) {
		t.Parallel()
		repo, err := New(context.Background(), s.URL)
		if err != nil {
			t.Fatal(err)
		}
		if e := repo.Auth(context.Background(), "succeed", "foo"); e != nil {
			t.Errorf("Unexpected error: %s", e)
		}
		if repo.user != "50230eec-ab2c-4e9e-96bc-57acee5ffae1" {
			t.Error("Failed to set user after auth")
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		t.Parallel()
		repo, err := New(context.Background(), s.URL)
		if err != nil {
			t.Fatal(err)
		}
		var msg string
		if e := repo.Auth(context.Background(), "fail", "foo"); e != nil {
			msg = e.Error()
		}
		if msg != "OAuth2 auth failed: Unauthorized" {
			t.Errorf("Unexpected error: %s", msg)
		}
	})
}

func TestLogout(t *testing.T) {
	s := mockServer(t)
	repo, err := New(context.Background(), s.URL)
	if err != nil {
		t.Fatal(err)
	}
	if e := repo.Auth(context.Background(), "succeed", "foo"); e != nil {
		t.Fatal(e)
	}
	if e := repo.Logout(context.Background()); e != nil {
		t.Errorf("Unexpected error: %s", e)
	}
	if repo.user != "" {
		t.Error("Failed to unset user")
	}
}

func TestCurrentUser(t *testing.T) {
	repo := &Repo{
		user: "bob",
	}
	if u, _ := repo.CurrentUser(); u != "bob" {
		t.Errorf("Got unexpected user: %s", u)
	}
}

var testDBCounter int32

func testClient(t *testing.T) kivikClient {
	c, err := localConnection()
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func testDB(t *testing.T) kivikDB {
	c := testClient(t)
	dbName := fmt.Sprintf("testdb-%x", atomic.AddInt32(&testDBCounter, 1))
	if e := c.CreateDB(context.Background(), dbName); e != nil {
		t.Fatal(e)
	}
	db, err := c.DB(context.Background(), dbName)
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestFetchUser(t *testing.T) {
	type fuTest struct {
		name     string
		db       kivikDB
		expected user
		status   int
	}
	tests := []fuTest{
		{
			name:   "NoUser",
			db:     testDB(t),
			status: kivik.StatusNotFound,
		},
		func() fuTest {
			db := testDB(t)
			rev, e := db.Put(context.Background(), currentUserDoc, map[string]string{"username": "foo"})
			if e != nil {
				t.Fatal(e)
			}

			return fuTest{
				name: "UserExists",
				db:   db,
				expected: user{
					ID:       currentUserDoc,
					Username: "foo",
					Rev:      rev,
				},
			}
		}(),
	}
	for _, test := range tests {
		func(test fuTest) {
			t.Run(test.name, func(t *testing.T) {
				repo := &Repo{
					state: test.db,
				}
				u, err := repo.fetchUser(context.Background())
				var status int
				if err != nil {
					status = kivik.StatusCode(err)
				}
				if status != test.status {
					t.Errorf("Unexpected error: %s", err)
				}
				if err != nil {
					return
				}
				if d := diff.AsJSON(test.expected, u); d != "" {
					t.Error(d)
				}
			})
		}(test)
	}
}

func TestSetUser(t *testing.T) {
	type suTest struct {
		name   string
		db     kivikDB
		status int
	}
	tests := []suTest{
		{
			name: "NewUser",
			db:   testDB(t),
		},
		{
			name: "ReplaceUser",
			db: func() kivikDB {
				db := testDB(t)
				if _, e := db.Put(context.Background(), currentUserDoc, map[string]string{"username": "alex"}); e != nil {
					t.Fatal(e)
				}
				return db
			}(),
		},
		{
			name: "SameUser",
			db: func() kivikDB {
				db := testDB(t)
				if _, e := db.Put(context.Background(), currentUserDoc, map[string]string{"username": "foo"}); e != nil {
					t.Fatal(e)
				}
				return db
			}(),
		},
	}
	for _, test := range tests {
		func(test suTest) {
			t.Run(test.name, func(t *testing.T) {
				repo := &Repo{
					state: test.db,
				}
				var status int
				err := repo.setUser(context.Background(), "foo")
				if err != nil {
					status = kivik.StatusCode(err)
				}
				if status != test.status {
					t.Errorf("Unexpected error: %s", err)
				}
				if err != nil {
					return
				}
				u, e := repo.fetchUser(context.Background())
				if e != nil {
					t.Fatal(e)
				}
				if u.Username != "foo" {
					t.Errorf("Unexpected result: %s", u.Username)
				}
			})
		}(test)
	}
}

func TestUserDB(t *testing.T) {
	type udTest struct {
		name       string
		repo       *Repo
		userDBName string
		err        string
	}
	tests := []udTest{
		{
			name: "Not logged in",
			repo: &Repo{},
			err:  "not logged in",
		},
		{
			name: "Connect failure",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				return &Repo{
					user:  "bob",
					local: local}
			}(),
			err: "database does not exist",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			udb, err := test.repo.userDB(context.Background())
			var msg string
			if err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s\n", msg)
			}
			if err != nil {
				return
			}
			stats, err := udb.Stats(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if stats.Name != test.userDBName {
				t.Errorf("Unexpected db name: %s", stats.Name)
			}
		})
	}
}

func TestBundleDB(t *testing.T) {
	type bdTest struct {
		name         string
		repo         *Repo
		bundle       *fb.Bundle
		bundleDBName string
		err          string
	}
	id, _ := fb.NewDbID("bundle", []byte{1, 2, 3, 4})
	testBundle := &fb.Bundle{ID: id}
	tests := []bdTest{
		{
			name:   "Not logged in",
			repo:   &Repo{},
			bundle: testBundle,
			err:    "not logged in",
		},
		{
			name:   "Invalid bundle",
			repo:   &Repo{user: "bob"},
			bundle: &fb.Bundle{},
			err:    "invalid bundle",
		},
		{
			name: "database not exist",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				return &Repo{
					user:  "bob",
					local: local}
			}(),
			bundle:       testBundle,
			bundleDBName: "bundle-aebagba",
		},
		{
			name: "success",
			repo: func() *Repo {
				local, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				dbName := testBundle.ID.String()
				if err := local.CreateDB(context.Background(), dbName); err != nil {
					t.Fatal(err)
				}
				return &Repo{
					user:  "bob",
					local: local}
			}(),
			bundle:       testBundle,
			bundleDBName: "bundle-aebagba",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bdb, err := test.repo.bundleDB(context.Background(), test.bundle)
			var msg string
			if err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s\n", msg)
			}
			if err != nil {
				return
			}
			stats, err := bdb.Stats(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if stats.Name != test.bundleDBName {
				t.Errorf("Unexpected db name: %s", stats.Name)
			}
		})
	}
}
