package anki

import (
	"html/template"
	"testing"
)

type clozeTest struct {
	TemplateID uint32
	Face       int
	Text       template.HTML
	Expected   template.HTML
}

func TestCloze(t *testing.T) {
	tests := []clozeTest{
		clozeTest{
			TemplateID: 0,
			Face:       0,
			Text:       "Quien mucho {{c1::abarca}} poco {{c2::aprieta}}.",
			Expected:   `Quien mucho <span class="cloze">[...]</span> poco aprieta.`,
		},
		clozeTest{
			TemplateID: 1,
			Face:       0,
			Text:       "Quien mucho {{c1::abarca}} poco {{c2::aprieta}}.",
			Expected:   `Quien mucho abarca poco <span class="cloze">[...]</span>.`,
		},
		clozeTest{
			TemplateID: 2,
			Face:       0,
			Text:       "Quien mucho {{c1::abarca}} poco {{c2::aprieta}}.",
			Expected:   "",
		},
		clozeTest{
			TemplateID: 0,
			Face:       1,
			Text:       "Quien mucho {{c1::abarca}} poco {{c2::aprieta}}.",
			Expected:   `Quien mucho <span class="cloze">abarca</span> poco aprieta.`,
		},
		clozeTest{
			TemplateID: 1,
			Face:       1,
			Text:       "Quien mucho {{c1::abarca}} poco {{c2::aprieta}}.",
			Expected:   `Quien mucho abarca poco <span class="cloze">aprieta</span>.`,
		},
		clozeTest{
			TemplateID: 2,
			Face:       1,
			Text:       "Quien mucho {{c1::abarca}} poco {{c2::aprieta}}.",
			Expected:   "",
		},
	}
	for _, test := range tests {
		fn := cloze(test.TemplateID, test.Face)
		result := fn(test.Text)
		if result != test.Expected {
			t.Errorf("Template %d, Face %d, %s\n\tExpected: %s\n\t  Actual: %s\n", test.TemplateID, test.Face, test.Text, test.Expected, result)
		}
	}
}
