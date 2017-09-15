// +build js,!cordova

package model

import "github.com/go-kivik/couchdb/chttp"

// setTransport does nothing for a standard web build.
func setTransport(_ *chttp.Client) {}
