package model

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/go-chi/chi"

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

func mockServer() *httptest.Server {
	r := chi.NewRouter()
	r.Get("/_session", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		resp := map[string]interface{}{
			"userCtx": map[string]interface{}{
				"name": "50230eec-ab2c-4e9e-96bc-57acee5ffae1",
			},
		}
		json.NewEncoder(w).Encode(resp)
	})
	r.Post("/_session", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var body struct {
			Provider string `json:"provider"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			panic(err)
		}
		if body.Provider != "succeed" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Cache-Control", "must-revalidate")
		w.Header().Add("Content-Type", "application/json")
		http.SetCookie(w, &http.Cookie{
			Name:     kivik.SessionCookieName,
			Value:    "NTAyMzBlZWMtYWIyYy00ZTllLTk2YmMtNTdhY2VlNWZmYWUxOjU5NTdFRjQzOjCE8hWO1JTg3XSt9FAZCjoIwzuT",
			Path:     "/",
			MaxAge:   10 * 60,
			HttpOnly: true,
		})
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"ok":    true,
			"name":  "50230eec-ab2c-4e9e-96bc-57acee5ffae1",
			"roles": []string{"user"},
		})
	})
	return httptest.NewServer(r)
}

func TestAuth(t *testing.T) {
	s := mockServer()
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
	s := mockServer()
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
