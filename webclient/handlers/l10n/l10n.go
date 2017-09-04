package l10n_handler

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"

	"golang.org/x/text/language"

	"github.com/flimzy/go-cordova"
	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/FlashbackSRS/flashback/l10n"
	"github.com/FlashbackSRS/flashback/util"
)

var jQuery = jquery.NewJQuery

const langTagAttr = "data-lt"
const localeAttr = "data-locale"

// Init initializes the localization engine.
func Init() *l10n.Set {
	langs := preferredLanguages()
	log.Debugf("PREFERRED LANGUAGES: %v\n", langs)
	set, err := l10n.New(langs, fetchTranslations)
	if err != nil {
		panic(err)
	}
	return set
}

func preferredLanguages() []language.Tag {
	var langs []language.Tag
	//	langs = append(langs, language.MustParse("es_MX"))
	if cordova.IsMobile() {
		var wg sync.WaitGroup
		wg.Add(1)
		nav := js.Global.Get("navigator")
		nav.Get("globalization").Call("getPreferredLanguage",
			func(l string) {
				defer wg.Done()
				if tag, err := language.Parse(l); err == nil {
					langs = append(langs, tag)
				}
			}, func() {
				defer wg.Done()
				// ignore any error
			})
		wg.Wait()
	}
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

// LocalizePage localizes all language tags on the page.
func LocalizePage(set *l10n.Set, h jqeventrouter.Handler) jqeventrouter.Handler {
	return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object, p url.Values) bool {
		go localize(set)
		return h.HandleEvent(event, ui, p)
	})
}

func localize(set *l10n.Set) {
	T, err := set.Tfunc()
	if err != nil {
		panic(err)
	}
	elements := jQuery("[" + langTagAttr + "]").Not("[" + localeAttr + "='" + set.Locale + "']")
	if elements.Length == 0 {
		log.Debugf("Nothing to localize, early exiting\n")
		return
	}
	log.Debugf("I think I have %d items to localize\n", elements.Length)
	for i := 0; i < elements.Length; i++ {
		elem := elements.Get(i)
		text := elem.Call("getAttribute", langTagAttr).String()
		if translated := T(text); translated != text {
			elem.Set("innerHTML", translated)
		} else {
			log.Debugf("%d. %s = %s\n", i, text, translated)
		}
	}
}
