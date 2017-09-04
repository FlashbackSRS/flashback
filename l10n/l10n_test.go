package l10n

import (
	"errors"
	"testing"

	"golang.org/x/text/language"
)

const testID = "foo"

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		prefLangs []language.Tag
		fetch     FetchCallback
		err       string
		initErr   string
		expected  string
	}{
		{
			name: "No callback",
			err:  "fetch callback required",
		},
		{
			name:     "No preference",
			fetch:    func(string) ([]byte, error) { return []byte(`[{"id":"foo","translation":"Foo"}]`), nil },
			expected: "Foo",
		},
		{
			name:      "Preference is default",
			prefLangs: []language.Tag{language.MustParse("en_US")},
			fetch:     func(_ string) ([]byte, error) { return []byte(`[{"id":"foo","translation":"Foo"}]`), nil },
			expected:  "Foo",
		},
		{
			name:      "Spanish preference",
			prefLangs: []language.Tag{language.MustParse("es_MX")},
			fetch:     func(_ string) ([]byte, error) { return []byte(`[{"id":"foo","translation":"Fóó"}]`), nil },
			expected:  "Fóó",
		},
		{
			name:      "Unsupported preference",
			prefLangs: []language.Tag{language.MustParse("de")},
			fetch:     func(_ string) ([]byte, error) { return []byte(`[{"id":"foo","translation":"Foo"}]`), nil },
			expected:  "Foo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			set, err := New(test.prefLangs, test.fetch)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != test.err {
				t.Errorf("Unexpected error: %s", errMsg)
			}
			if err != nil {
				return
			}
			set.initWG.Wait()
			T, err := set.Tfunc()
			var initErr string
			if err != nil {
				initErr = err.Error()
			}
			if test.initErr != initErr {
				t.Errorf("Unexpected init error: %s", initErr)
			}
			if err != nil {
				return
			}
			result := T(testID)
			if test.expected != result {
				t.Errorf("Unexpected translation: %s", result)
			}
		})
	}
}

func TestLoadDictionary(t *testing.T) {
	tests := []struct {
		name     string
		locale   string
		fetch    FetchCallback
		expected string
		err      string
	}{
		{
			name:   "fetch error",
			locale: "foo",
			fetch: func(_ string) ([]byte, error) {
				return nil, errors.New("fetch error")
			},
			err: "fetch error",
		},
		{
			name:   "invalid translation data",
			locale: "foo",
			fetch: func(_ string) ([]byte, error) {
				return []byte("foo"), nil
			},
			err: `no language found in "foo.all.json"`,
		},
		{
			name:   "success",
			locale: "en-us",
			fetch: func(_ string) ([]byte, error) {
				return []byte(`[{"id":"foo","translation":"Foo"}]`), nil
			},
			expected: "Foo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			T, err := loadDictionary(test.locale, test.fetch)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if test.err != errMsg {
				t.Errorf("Unexpected error: %s", errMsg)
			}
			if err != nil {
				return
			}
			result := T(testID)
			if result != test.expected {
				t.Errorf("Unexpected translation: %s", result)
			}
		})
	}
}
