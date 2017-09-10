// +build cordova

package l10n_handler

import (
	"errors"
	"sync"

	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/console"
)

func preferredLanguages() ([]string, error) {
	var langs []string
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	nav := js.Global.Get("navigator")
	nav.Get("globalization").Call("getPreferredLanguage",
		func(l string) {
			langs = append(langs, l)
			wg.Done()
		}, func(e *js.Object) {
			err = errors.New(e.Call("toString").String())
			console.Log(e)
			wg.Done()
		})
	wg.Wait()
	return langs, err
}
