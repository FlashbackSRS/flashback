// +build cordova

package l10n_handler

import (
	"sync"

	"honnef.co/go/js/console"

	"github.com/gopherjs/gopherjs/js"
	"golang.org/x/text/language"
)

func preferredLanguages() []language.Tag {
	var langs []language.Tag
	var wg sync.WaitGroup
	wg.Add(1)
	nav := js.Global.Get("navigator")
	nav.Get("globalization").Call("getPreferredLanguage",
		func(l string) {
			defer wg.Done()
			if tag, err := language.Parse(l); err == nil {
				langs = append(langs, tag)
			}
		}, func(e *js.Object) {
			console.Log(e)
			defer wg.Done()
		})
	wg.Wait()
	return langs
}
