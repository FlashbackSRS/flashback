package anki

import (
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"

	"github.com/FlashbackSRS/flashback/model"
)

// Cloze is the controller for the Anki Cloze model.
type Cloze struct {
	*Basic
}

var _ model.ModelController = &Cloze{}
var _ model.FuncMapper = &Cloze{}

// Type returns the string "anki-cloze", to identify this model handler's type.
func (m *Cloze) Type() string {
	return "anki-cloze"
}

// FuncMap returns a function map for Cloze templates.
func (m *Cloze) FuncMap(card *model.Card, face int) template.FuncMap {
	var templateID uint32
	if card != nil {
		// Need to do this check, because card may be nil during template parsing
		templateID = card.TemplateID()
	}
	funcMap := map[string]interface{}{
		"cloze": cloze(templateID, face),
	}
	for k, v := range defaultFuncMap {
		funcMap[k] = v
	}
	return funcMap
}

const span = `<span class="cloze">%s</span>`

var placeholder = fmt.Sprintf(span, "[...]")

var clozeRE = regexp.MustCompile(`{{c(\d+)::(.*?)}}`)

func cloze(templateID uint32, face int) func(template.HTML) template.HTML {
	return func(text template.HTML) template.HTML {
		str := string(text)
		var found bool
		for _, match := range clozeRE.FindAllStringSubmatch(str, -1) {
			fieldNo, _ := strconv.Atoi(match[1])
			// Subtract one, because templateID is 0-indexed, but {{cN:...}} fields are 1-indexed
			index := fieldNo - 1
			if index == int(templateID) {
				found = true
				if face == 0 {
					str = strings.Replace(str, match[0], placeholder, -1)
				} else {
					str = strings.Replace(str, match[0], fmt.Sprintf(span, match[2]), -1)
				}
			} else {
				str = strings.Replace(str, match[0], match[2], -1)
			}
		}
		if found {
			return template.HTML(str)
		}
		return ""
	}
}
