package l10n

import (
	"errors"
	"strings"
	"sync"

	"github.com/flimzy/log"
	"github.com/nicksnyder/go-i18n/i18n/bundle"

	"golang.org/x/text/language"
)

var matcher = language.NewMatcher([]language.Tag{
	language.MustParse("en_US"),
	language.MustParse("es_MX"),
})

// Set represents a language set.
type Set struct {
	Locale        string
	initWG        sync.WaitGroup
	tfunc         bundle.TranslateFunc
	fallbackTfunc bundle.TranslateFunc
	err           error
}

// FetchCallback receives the name of a locale, and must return the translation
// rules (in JSON) for that locale.
type FetchCallback func(locale string) ([]byte, error)

// New initializes a new language set.
func New(preferredLanguages []language.Tag, fetch FetchCallback) (*Set, error) {
	if fetch == nil {
		return nil, errors.New("fetch callback required")
	}
	tag, idx, conf := matcher.Match(preferredLanguages...)
	log.Debugf("Selected language %s (preference choice %d with %s confidence)", tag, idx, conf)
	s := &Set{
		Locale: strings.ToLower(tag.String()),
	}
	s.initWG.Add(1)
	go s.init(tag, fetch)
	return s, nil
}

func (s *Set) init(tag language.Tag, fetch FetchCallback) {
	defer s.initWG.Done()
	if s.Locale != "en-us" {
		s.tfunc, s.err = loadDictionary(s.Locale, fetch)
		if s.err != nil {
			return
		}
	}
	s.fallbackTfunc, s.err = loadDictionary("en-us", fetch)
}

func loadDictionary(locale string, fetch FetchCallback) (bundle.TranslateFunc, error) {
	translations, err := fetch(locale)
	if err != nil {
		return nil, err
	}
	bdl := bundle.New()
	if e := bdl.ParseTranslationFileBytes(locale+".all.json", translations); e != nil {
		return nil, e
	}
	return bdl.Tfunc(locale, locale)
}

// Tfunc returns a translation function.
func (s *Set) Tfunc() (bundle.TranslateFunc, error) {
	s.initWG.Wait()
	if s.err != nil {
		return nil, s.err
	}
	return func(id string, args ...interface{}) string {
		if s.tfunc != nil {
			if result := s.tfunc(id, args...); result != id {
				return result
			}
			log.Debugf("No result looking up tag '%s'", id)
		}
		if result := s.fallbackTfunc(id, args...); result != id {
			return result
		}
		log.Debugf("No result looking up fallback tag '%s'", id)
		return id
	}, nil
}
