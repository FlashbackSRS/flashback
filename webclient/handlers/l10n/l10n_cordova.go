// +build cordova

package l10n_handler

import (
	"errors"
	"fmt"
	"sync"

	"github.com/FlashbackSRS/flashback/l10n"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/console"
)

type source struct{}

var _ l10n.Source = &source{}

func langSource(_ string) l10n.Source {
	return &source{}
}

func (src *source) Languages() ([]string, error) {
	var langs []string
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	globalization := js.Global.Get("navigator").Get("globalization")
	if globalization == js.Undefined {
		return nil, errors.New("cannot find globalization object; is cordova-plugin-globalization loaded?")
	}
	globalization.Call("getPreferredLanguage",
		func(l *js.Object) {
			langs = append(langs, l.Get("value").String())
			wg.Done()
		}, func(e *js.Object) {
			err = errors.New(e.Call("toString").String())
			console.Log(e)
			wg.Done()
		})
	wg.Wait()
	return langs, err
}

func (src *source) FetchLanguage(lang string) ([]byte, error) {
	if js.Global.Get("resolveLocalFileSystemURL") == js.Undefined {
		return nil, errors.New("resolveLocalFileSystemURL function undefined; is cordova-plugin-file loaded?")
	}
	var content []byte
	var err error
	var wg sync.WaitGroup
	wg.Add(1)
	path := appDir() + "translations/" + lang + ".all.json"
	log.Debugf("trying to read file: %s\n", path)
	errHandler := func(e *js.Object) {
		err = fileError(path, e)
		wg.Done()
	}
	js.Global.Call("resolveLocalFileSystemURL", path, func(fileEntry *js.Object) {
		fileEntry.Call("file", func(file *js.Object) {
			reader := js.Global.Get("FileReader").New()
			reader.Set("onloadend", func() {
				if e := reader.Get("error"); e != nil {
					console.Log(e)
				}
				content = []byte(reader.Get("result").String())
				wg.Done()
			})
			reader.Call("readAsText", file)
		}, errHandler)
	}, errHandler)
	wg.Wait()
	return content, err
}

func appDir() string {
	return js.Global().Get("cordova").Get("file").Get("applicationDirectory").String() + "www/"
}

const (
	cordova_NOT_FOUND_ERR = iota
	cordova_SECURITY_ERR
	cordova_ABORT_ERR
	cordova_NOT_READABLE_ERR
	cordova_ENCODING_ERR
	cordova_NO_MODIFICATION_ALLOWED_ERR
	cordova_INVALID_STATE_ERR
	cordova_SYNTAX_ERR
	cordova_INVALID_MODIFICATION_ERR
	cordova_QUOTA_EXCEEDED_ERR
	cordova_TYPE_MISMATCH_ERR
	cordova_PATH_EXISTS_ERR
)

func fileError(filename string, e *js.Object) error {
	var msg string
	switch e.Get("code").Int() {
	case cordova_NOT_FOUND_ERR:
		msg = "not found"
	case cordova_SECURITY_ERR:
		msg = "security error"
	case cordova_ABORT_ERR:
		msg = "aborted"
	case cordova_NOT_READABLE_ERR:
		msg = "not readable"
	case cordova_ENCODING_ERR:
		msg = "encoding error"
	case cordova_NO_MODIFICATION_ALLOWED_ERR:
		msg = "no modification allowed"
	case cordova_INVALID_STATE_ERR:
		msg = "invalid state"
	case cordova_SYNTAX_ERR:
		msg = "syntax error"
	case cordova_INVALID_MODIFICATION_ERR:
		msg = "invalid modification"
	case cordova_QUOTA_EXCEEDED_ERR:
		msg = "quota exceeded"
	case cordova_TYPE_MISMATCH_ERR:
		msg = "type mismatch"
	case cordova_PATH_EXISTS_ERR:
		msg = "path exists"
	default:
		msg = "unknown error"
	}
	return fmt.Errorf("%s: %s", filename, msg)
}
