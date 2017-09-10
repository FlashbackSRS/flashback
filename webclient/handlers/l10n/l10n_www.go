// +build !cordova

package l10n_handler

import (
	"io/ioutil"
	"net/http"

	"github.com/FlashbackSRS/flashback/l10n"
	"github.com/gopherjs/gopherjs/js"
)

type source struct {
	baseURL string
}

var _ l10n.Source = &source{}

func langSource(baseURL string) l10n.Source {
	return &source{
		baseURL: baseURL,
	}
}

func (src *source) FetchLanguage(lang string) ([]byte, error) {
	resp, err := http.Get(src.baseURL + "translations/" + lang + ".all.json")
	if err != nil {
		return []byte{}, err
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return content, nil
}

func (src *source) Languages() ([]string, error) {
	var langs []string

	if languages := js.Global.Get("navigator").Get("languages"); languages != nil {
		for i := 0; i < languages.Length(); i++ {
			langs = append(langs, languages.Index(i).String())
		}
	}
	langs = append(langs, js.Global.Get("navigator").Get("language").String())
	return langs, nil
}
