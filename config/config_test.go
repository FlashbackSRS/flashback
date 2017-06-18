package config

import (
	"testing"

	"github.com/flimzy/diff"
)

func TestNewFromJSON(t *testing.T) {
	c, err := NewFromJSON([]byte(`{"foo":"bar"}`))
	if err != nil {
		t.Fatal(err)
	}
	expected := &Conf{
		c: map[string]string{"foo": "bar"},
	}
	if d := diff.Interface(expected, c); d != "" {
		t.Error(d)
	}
}

func TestNewFromInvalidJSON(t *testing.T) {
	_, err := NewFromJSON([]byte(`invalid json`))
	if err == nil {
		t.Fatal("Expected an error")
	}
}

func TestNew(t *testing.T) {
	c := New(map[string]string{"foo": "bar"})
	expected := &Conf{
		c: map[string]string{"foo": "bar"},
	}
	if d := diff.Interface(expected, c); d != "" {
		t.Error(d)
	}
}

func TestIsSet(t *testing.T) {
	c := New(map[string]string{"foo": "bar"})
	if !c.IsSet("foo") {
		t.Error("Expected 'foo' to be set")
	}
	if c.IsSet("bar") {
		t.Error("Expected 'bar' not to be set")
	}
}

func TestGetString(t *testing.T) {
	c := New(map[string]string{"foo": "bar"})
	if c.GetString("foo") != "bar" {
		t.Error("Expected 'bar' to be returned")
	}
	if c.GetString("bar") != "" {
		t.Error("Expected '' to be returned")
	}
}
