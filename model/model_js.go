// +build js

package model

import (
	"context"
	"net/http"

	cordova "github.com/flimzy/go-cordova"
	"github.com/flimzy/kivik"
	"github.com/go-kivik/couchdb/chttp"
	_ "github.com/go-kivik/pouchdb" // PouchDB driver
)

func localConnection() (kivikClient, error) {
	c, err := kivik.New(context.Background(), "pouch", "")
	if err != nil {
		return nil, err
	}
	return wrapClient(c), nil
}

func remoteConnection(dsn string) (kivikClient, error) {
	c, err := kivik.New(context.Background(), "pouch", dsn)
	if err != nil {
		return nil, err
	}
	return wrapClient(c), nil
}

// setTransport exists for the benefit of Cordova, which appears to ignore
// Set-Cookie headers in Fetch responses; so this explicitly uses the XHR
// interface.
func setTransport(client *chttp.Client) {
	if cordova.IsMobile() {
		client.Transport = &http.XHRTransport{}
	}
}
