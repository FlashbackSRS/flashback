// +build js

package model

import (
	"net/http/httptest"
	"testing"

	"github.com/gopherjs/gopherjs/js"
)

const env = "js"

func init() {
	newPouch := js.Global.Get("PouchDB").Call("defaults", map[string]interface{}{
		"db": js.Global.Call("require", "memdown"),
	})
	js.Global.Set("PouchDB", newPouch)
}

func mockServer(t *testing.T) *httptest.Server {
	t.Skip("Cannot run server in GopherJS")
	return nil
}
