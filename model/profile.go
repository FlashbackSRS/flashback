package model

import (
	"time"

	"github.com/flimzy/log"
)

func profile(name string) func() {
	start := time.Now()
	log.Debugf("Starting %s", name)
	return func() {
		finish := time.Now()
		log.Debugf("Finished %s (%v)", name, finish.Sub(start))
	}
}
