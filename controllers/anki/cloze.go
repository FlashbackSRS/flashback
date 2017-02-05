package anki

import (
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"
)

const span = `<span class="cloze">%s</span>`

var placeholder = fmt.Sprintf(span, "[...]")

var clozeRE = regexp.MustCompile(`{{c(\d+)::(.*?)}}`)

func cloze(face int, cardNo int, text string) template.HTML {
	var found bool
	for _, match := range clozeRE.FindAllStringSubmatch(text, -1) {
		fieldNo, _ := strconv.Atoi(match[1])
		if fieldNo == cardNo {
			found = true
			if face == 0 {
				text = strings.Replace(text, match[0], placeholder, -1)
			} else {
				text = strings.Replace(text, match[0], fmt.Sprintf(span, match[2]), -1)
			}
		} else {
			text = strings.Replace(text, match[0], match[2], -1)
		}
	}
	if found {
		return template.HTML(text)
	}
	return ""
}
