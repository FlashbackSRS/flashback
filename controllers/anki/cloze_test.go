package anki

import (
	"bytes"
	"fmt"
	"html/template"
	"testing"
)

type clozeTest struct {
	CardNo   int
	Face     int
	Text     string
	Expected template.HTML
}

func TestCloze(t *testing.T) {
	tests := []clozeTest{
		clozeTest{
			CardNo:   1,
			Face:     0,
			Text:     "Quien mucho {{c1::abarca}} poco {{c2::aprieta}}.",
			Expected: `Quien mucho <span class="cloze">[...]</span> poco aprieta.`,
		},
		clozeTest{
			CardNo:   2,
			Face:     0,
			Text:     "Quien mucho {{c1::abarca}} poco {{c2::aprieta}}.",
			Expected: `Quien mucho abarca poco <span class="cloze">[...]</span>.`,
		},
		clozeTest{
			CardNo:   3,
			Face:     0,
			Text:     "Quien mucho {{c1::abarca}} poco {{c2::aprieta}}.",
			Expected: "",
		},
		clozeTest{
			CardNo:   1,
			Face:     1,
			Text:     "Quien mucho {{c1::abarca}} poco {{c2::aprieta}}.",
			Expected: `Quien mucho <span class="cloze">abarca</span> poco aprieta.`,
		},
		clozeTest{
			CardNo:   2,
			Face:     1,
			Text:     "Quien mucho {{c1::abarca}} poco {{c2::aprieta}}.",
			Expected: `Quien mucho abarca poco <span class="cloze">aprieta</span>.`,
		},
		clozeTest{
			CardNo:   3,
			Face:     1,
			Text:     "Quien mucho {{c1::abarca}} poco {{c2::aprieta}}.",
			Expected: "",
		},
	}
	for _, test := range tests {
		result := cloze(test.Face, test.CardNo, test.Text)
		if result != test.Expected {
			t.Errorf("Card %d, Face %d, %s\n\tExpected: %s\n\t  Actual: %s\n", test.CardNo, test.Face, test.Text, test.Expected, result)
		}
	}
}

type card struct{}

func (c *card) ModelID() int {
	return 1
}

func TestClozeTemplate(t *testing.T) {
	tmpl, err := template.New("template").Funcs(map[string]interface{}{"cloze": cloze}).Parse("{{cloze .Face .Card.ModelID .Foo}}")
	if err != nil {
		t.Fatalf("error parsing template: %s\n", err)
	}
	ctx := map[string]interface{}{
		"Face": 1,
		"Card": &card{},
		"Foo":  "Foo {{c1::bar}} baz",
	}
	result := new(bytes.Buffer)
	if err := tmpl.Execute(result, ctx); err != nil {
		t.Fatalf("Error executing template: %s\n", err)
	}
	fmt.Printf("Reesult = %s\n", result.String())
}
