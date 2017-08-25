// +build !js

package model

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/go-chi/chi"
)

const env = "go"

func mockServer(_ *testing.T) *httptest.Server {
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
