// +build !cordova

package main

import (
	"net/url"
	"strings"
	"sync"
)

func initCordova(_ *sync.WaitGroup) {
	// nothing to do here
	return
}

// urlPrefix returns the URL prefix for routing purposes.
func urlPrefix(baseURL string) string {
	parsed, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}
	return strings.TrimSuffix(parsed.Path, "/")
}
