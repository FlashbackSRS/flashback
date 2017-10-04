package index

import (
	"html/template"
)

//go:generate go-bindata -pkg index -nocompress -prefix files -o data.go files

const templateFilename = "deck-list.html"

var parsedTemplate *template.Template

func deckListTemplate() (*template.Template, error) {
	if parsedTemplate == nil {
		data, err := Asset(templateFilename)
		if err != nil {
			return nil, err
		}
		tmpl, err := template.New(templateFilename).Parse(string(data))
		if err != nil {
			return nil, err
		}
		parsedTemplate = tmpl
	}
	return parsedTemplate, nil
}
