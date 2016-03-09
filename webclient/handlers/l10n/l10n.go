package l10n_handler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/nicksnyder/go-i18n/i18n/bundle"
	"golang.org/x/text/language"

	"github.com/flimzy/jqeventrouter"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"

	"github.com/flimzy/flashback/util"
	"github.com/flimzy/go-cordova"
)

var localeName string
var jQuery = jquery.NewJQuery
var initDone <-chan struct{}

const langTagAttr = "data-lt"
const localeAttr = "data-locale"

var translate *bundle.TranslateFunc
var translateFallback *bundle.TranslateFunc

func MobileInit() {
	initChan := make(chan struct{})
	initDone = initChan
	go func() {
		defer close(initChan)
		langs := preferredLanguages()
		fmt.Printf("PREFERRED LANGUAGES: %v\n", langs)
		matcher := language.NewMatcher([]language.Tag{
			language.MustParse("en_US"),
			language.MustParse("es_MX"),
		})
		tag, idx, conf := matcher.Match(langs...)
		fmt.Printf("I think we'll go with %s (%d th place choice with %s confidence)\n", tag, idx, conf)
		if locale := tag.String(); locale != "en-us" {
			t, err := loadDictionary(locale)
			if err != nil {
				panic(fmt.Sprintf("Cannot load '%s': %s", locale, err))
			}
			translate = t
		}
		if t, err := loadDictionary("en-us"); err != nil {
			panic(fmt.Sprintf("Cannot load 'en-us': %s", err))
		} else {
			translateFallback = t
		}
	}()
}

func loadDictionary(locale string) (*bundle.TranslateFunc, error) {
	locale = strings.ToLower(locale)
	translations, err := fetchTranslations(locale)
	if err != nil {
		return nil, err
	}
	bdl := bundle.New()
	if err := bdl.ParseTranslationFileBytes(locale+".all.json", translations); err != nil {
		return nil, err
	}
	t, err := bdl.Tfunc(locale, locale) // stupid API, requires a second parameter
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func T(id string, args ...interface{}) string {
	if translate != nil {
		t := *translate
		// If dictionary == nil, it means the requested language is the default
		if result := t(id, args...); result != id {
			return result
		}
		fmt.Printf("No result looking up tag '%s'\n", id)
		// TODO: Log the missing translation
	}
	t := *translateFallback
	if result := t(id, args...); result != id {
		return result
	}
	fmt.Printf("No result looking up fallback tag '%s'\n", id)
	// TODO: *really* log the missing translation!
	return id
}

func preferredLanguages() []language.Tag {
	var langs []language.Tag
	//	langs = append(langs, language.MustParse("es_MX"))
	nav := js.Global.Get("navigator")
	if cordova.IsMobile() {
		var wg sync.WaitGroup
		wg.Add(1)
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

func LocalizePage(h jqeventrouter.Handler) jqeventrouter.Handler {
	return jqeventrouter.HandlerFunc(func(event *jquery.Event, ui *js.Object, p url.Values) bool {
		go localize()
		return h.HandleEvent(event, ui, p)
	})
}

func localize() {
	<-initDone // Make sure the dictionary is ready before we try translating
	elements := jQuery("[" + langTagAttr + "]").Not("[" + localeAttr + "='" + localeName + "']")
	if elements.Length == 0 {
		fmt.Printf("Nothing to localize, early exiting\n")
		return
	}
	fmt.Printf("I think I have %d items to localize\n", elements.Length)
	for i := 0; i < elements.Length; i++ {
		elem := elements.Get(i)
		text := elem.Call("getAttribute", langTagAttr).String()
		if translated := T(text); translated != text {
			elem.Set("innerHTML", translated)
		} else {
			fmt.Printf("%d. %s = %s\n", i, text, translated)
		}
	}
}
