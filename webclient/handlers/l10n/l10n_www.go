// +build !cordova

package l10n_handler

import (
	"github.com/gopherjs/gopherjs/js"
)

func preferredLanguages() ([]string, error) {
	var langs []string

	if languages := js.Global.Get("navigator").Get("languages"); languages != nil {
		for i := 0; i < languages.Length(); i++ {
			langs = append(langs, languages.Index(i).String())
		}
	}
	langs = append(langs, js.Global.Get("navigator").Get("language").String())
	return langs, nil
}
