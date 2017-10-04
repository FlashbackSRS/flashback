package index

import "testing"

func TestDeckListTemplate(t *testing.T) {
	tmpl, err := deckListTemplate()
	if err != nil {
		t.Fatal(err)
	}
	expectedName := templateFilename
	if tmpl.Name() != expectedName {
		t.Errorf("Unexpected name: %s", tmpl.Name())
	}
}
