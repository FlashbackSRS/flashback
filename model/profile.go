package model

import (
	"fmt"
	"time"

	"github.com/flimzy/log"
)

func profile(format string, args ...interface{}) func() {
	name := fmt.Sprintf(format, args...)
	start := time.Now()
	log.Debugf("Starting %s", name)
	return func() {
		finish := time.Now()
		log.Debugf("Finished %s (%v)", name, finish.Sub(start))
	}
}
