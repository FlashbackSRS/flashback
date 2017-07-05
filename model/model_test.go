package model

import (
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"

	"golang.org/x/net/context"
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
		if msg != "Unauthorized" {
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
	if u := repo.CurrentUser(); u != "bob" {
		t.Errorf("Got unexpected user: %s", u)
	}
}

func TestFetchUser(t *testing.T) {
	type fuTest struct {
		name     string
		db       *kivik.DB
		expected user
		status   int
	}
	tests := []fuTest{
		{
			name: "NoUser",
			db: func() *kivik.DB {
				c, err := localConnection()
				if err != nil {
					t.Fatal(err)
				}
				if e := c.CreateDB(context.Background(), "a"); e != nil {
					t.Fatal(e)
				}
				db, err := c.DB(context.Background(), "a")
				if err != nil {
					t.Fatal(err)
				}
				return db
			}(),
			status: kivik.StatusNotFound,
		},
		func() fuTest {
			c, err := localConnection()
			if err != nil {
				t.Fatal(err)
			}
			if e := c.CreateDB(context.Background(), "b"); e != nil {
				t.Fatal(e)
			}
			db, err := c.DB(context.Background(), "b")
			if err != nil {
				t.Fatal(err)
			}
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

// func TestStoreUser(t *testing.T) {
// 	type suTest struct {
// 		name    string
// 		localDB *kivik.DB
// 		err     string
// 	}
// 	tests := []suTest{}
// 	for _, test := range tests {
// 		func(test suTest) {
// 			t.Run(test.name, func(t *testing.T) {
//
// 			})
// 		}(test)
// 	}
// }
