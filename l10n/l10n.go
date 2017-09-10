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

// Source is a language information source.
type Source interface {
	// FetchLanguage receives the name of a language and must return the
	// translation dictionary (in JSON) for that language.
	FetchLanguage(lang string) ([]byte, error)
	// Languages returns a list of language tags, in order of user preference.
	Languages() ([]string, error)
}

// New initializes a new language set.
func New(src Source) (*Set, error) {
	if src == nil {
		return nil, errors.New("src required")
	}
	s := &Set{}
	s.initWG.Add(1)
	go s.init(src)
	return s, nil
}

func (s *Set) init(src Source) {
	defer log.Debug("l10n init complete")
	defer s.initWG.Done()
	tags, err := src.Languages()
	if err != nil {
		log.Debugf("langs cb failed: %s", err)
		s.err = err
		return
	}
	log.Debugf("Preferred languages: %v", tags)
	langs := make([]language.Tag, 0, len(tags))
	for _, tag := range tags {
		if tag, err := language.Parse(tag); err == nil {
			langs = append(langs, tag)
		}
	}
	tag, idx, conf := matcher.Match(langs...)
	log.Debugf("Selected language %s (preference choice %d with %s confidence)", tag, idx, conf)
	s.Locale = strings.ToLower(tag.String())
	if s.Locale != "en-us" {
		s.tfunc, s.err = loadDictionary(s.Locale, src)
		if s.err != nil {
			log.Debugf("load dict failed: %s", s.err)
			return
		}
	}
	s.fallbackTfunc, s.err = loadDictionary("en-us", src)
	if s.err != nil {
		log.Debugf("load dict 2 failed: %s", s.err)
	}
}

func loadDictionary(locale string, src Source) (bundle.TranslateFunc, error) {
	translations, err := src.FetchLanguage(locale)
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
