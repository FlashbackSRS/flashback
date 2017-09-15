// +build js,cordova

package model

import "github.com/go-kivik/couchdb/chttp"

// setTransport exists for the benefit of Cordova, which appears to ignore
// Set-Cookie headers in Fetch responses; so this explicitly uses the XHR
// interface.
func setTransport(client *chttp.Client) {
	client.Transport = &http.XHRTransport{}
}
