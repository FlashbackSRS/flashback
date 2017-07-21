// +build js

package model

import (
	"context"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/pouchdb" // PouchDB driver
)

func localConnection() (*kivik.Client, error) {
	return kivik.New(context.Background(), "pouch", "")
}

func remoteConnection(dsn string) (*kivik.Client, error) {
	return kivik.New(context.Background(), "pouch", dsn)
}
