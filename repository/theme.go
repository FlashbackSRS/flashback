package repo

import (
	"fmt"
	"html/template"

	"github.com/flimzy/log"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback-model"
)

// Theme is a wrapper around a fb.Theme object
type Theme struct {
	*fb.Theme
}

// Model is a wrapper around a fb.Model object
type Model struct {
	*fb.Model
}

// GenerateTemplate returns a string representing the Model's rendered template.
func (m *Model) GenerateTemplate() (*template.Template, error) {
	log.Debugf("model: %v", m)
	mainTemplate := fmt.Sprintf("$template.%d.html", m.ID)
	if _, ok := m.Files.GetFile(mainTemplate); !ok {
		return nil, errors.New("Main template not found in model")
	}
	templates, err := extractTemplateFiles(m.Files)
	if err != nil {
		return nil, errors.Wrap(err, "Error with Model attachments")
	}
	log.Debugf("templates = %v\n", templates)
	tmpl2, err := extractTemplateFiles(m.Theme.Files)
	if err != nil {
		return nil, errors.Wrap(err, "Error with Theme attachments")
	}
	log.Debugf("templates = %v\n", templates)
	log.Debugf("tmpl2 = %v\n", tmpl2)
	for k, v := range tmpl2 {
		templates[k] = v
	}
	log.Debugf("templates = %v\n", templates)
	// Rename to match the masterTemplate expectation
	templates["template.html"] = templates[mainTemplate]
	delete(templates, mainTemplate)
	tmpl, err := template.New("template").Parse(masterTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing master template")
	}
	for filename, t := range templates {
		log.Debugf("Defining template '%s'", filename)
		content := fmt.Sprintf("{{define \"%s\"}}%s{{end}}", filename, t)
		if _, err := tmpl.Parse(content); err != nil {
			return nil, errors.Wrapf(err, "Error parsing template file `%s`", filename)
		}
	}
	return tmpl, nil
}

var templateTypes = map[string]struct{}{
	fb.TemplateContentType: struct{}{},
	"text/css":             struct{}{},
	"script/javascript":    struct{}{},
}

func extractTemplateFiles(v *fb.FileCollectionView) (map[string]string, error) {
	templates := make(map[string]string)
	for _, filename := range v.FileList() {
		log.Debugf("Filename: %s\n", filename)
		att, ok := v.GetFile(filename)
		if !ok {
			return nil, errors.Errorf("Error fetching expected file '%s' from Model", filename)
		}
		log.Debugf("ContentType: %s\n", att.ContentType)
		if _, ok := templateTypes[att.ContentType]; ok {
			templates[filename] = string(att.Content)
		}
	}
	return templates, nil
}

var masterTemplate = `
<!DOCTYPE html>
<html>
<head>
	<title>FB Card</title>
	<base href="{{ .BaseURI }}">
	<meta charset="UTF-8">
	<meta http-equiv="Content-Security-Policy"
		content="script-src 'unsafe-inline' {{ .BaseURI }}">
<script type="text/javascript">
'use strict';
var FB = {
	iframeID: '{{ .IframeID }}',
	card: {{ .Card }},
	note: {{ .Note }},
	model: {{ .Model }},
};
</script>
<script type="text/javascript" src="js/cardframe.js"></script>
<script type="text/javascript">{{ block "script.js" .Fields }}{{end}}</script>
<style>{{ block "style.css" .Fields }}{{end}}</style>
</head>
<body class="card">{{ block "template.html" .Fields }}{{end}}</body>
</html>
`
