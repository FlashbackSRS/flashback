// +build !cordova

package l10n_handler

import (
	"io/ioutil"
	"net/http"

	"github.com/FlashbackSRS/flashback/util"
)

func fetchTranslations(lang string) ([]byte, error) {
	resp, err := http.Get(util.BaseURI() + "translations/" + lang + ".all.json")
	if err != nil {
		return []byte{}, err
	}
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}
	return content, nil
}
