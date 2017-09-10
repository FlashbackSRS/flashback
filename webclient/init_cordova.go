// +build cordova

package main

import (
	"sync"

	"github.com/flimzy/log"
)

func initCordova(wg *sync.WaitGroup) {
	log.Debug("Initializing Cordova extensions\n")
	wg.Add(1)
	document.Call("addEventListener", "deviceready", func() {
		// TODO: Don't defer; and perhaps even pass wg.Done directly to the Call method?
		defer wg.Done()
	}, false)
	log.Debug("Cordova init complete\n")
}
