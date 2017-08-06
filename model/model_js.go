// +build js

package model

import (
	"context"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/pouchdb" // PouchDB driver
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
