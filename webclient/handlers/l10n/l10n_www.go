// +build !cordova

package l10n_handler

import (
	"github.com/gopherjs/gopherjs/js"
	"golang.org/x/text/language"
)

func preferredLanguages() []language.Tag {
	var langs []language.Tag

	if languages := js.Global.Get("navigator").Get("languages"); languages != nil {
		for i := 0; i < languages.Length(); i++ {
			if tag, err := language.Parse(languages.Index(i).String()); err == nil {
				langs = append(langs, tag)
			}
		}
	}
	if tag, err := language.Parse(js.Global.Get("navigator").Get("language").String()); err == nil {
		langs = append(langs, tag)
	}
	return langs
}
