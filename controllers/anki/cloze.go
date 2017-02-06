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
