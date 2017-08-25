package anki

import "testing"

func TestAnkiBasicConvertQuery(t *testing.T) {
	expected := &basicQuery{
		TypedAnswers: map[string]string{"foo": "bar"},
	}
	result := convertQuery(expected)
	if result != expected {
		t.Errorf("Unexpected result")
	}
}
