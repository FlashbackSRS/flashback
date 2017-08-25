// +build js

package anki

import (
	"testing"

	"github.com/flimzy/diff"
	"github.com/gopherjs/gopherjs/js"
)

func TestAnkiBasicConvertQueryJS(t *testing.T) {
	expected := &basicQuery{
		Submit:       "foo",
		TypedAnswers: map[string]string{"foo": "bar"},
	}
	query := js.Global.Get("Object").New()
	query.Set("type:foo", "bar")
	query.Set("ignore", "baz")
	query.Set("submit", "foo")
	result := convertQuery(query)
	if d := diff.Interface(expected, result); d != nil {
		t.Error(d)
	}
}
