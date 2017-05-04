package repo

import (
	"fmt"
	"html/template"
	"strings"

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

// FuncMap returns the model controller's FuncMap, if any.
func (c *PouchCard) FuncMap(face int) (template.FuncMap, error) {
	mc, err := c.getModelController()
	if err != nil {
		return nil, err
	}
	if funcMapper, ok := mc.(FuncMapper); ok {
		return funcMapper.FuncMap(c, face), nil
	}
	return nil, nil
}

// FuncMap returns a placeholder template.FuncMap for use when parsing templates.
func (m *Model) FuncMap() (template.FuncMap, error) {
	mc, err := m.getController()
	if err != nil {
		return nil, err
	}
	if funcMapper, ok := mc.(FuncMapper); ok {
		return funcMapper.FuncMap(nil, 0), nil
	}
	return nil, nil
}

// GenerateTemplate returns a string representing the Model's rendered template.
func (m *Model) GenerateTemplate() (*template.Template, error) {
	mainTemplate := fmt.Sprintf("$template.%d.html", m.ID)
	if _, ok := m.Files.GetFile(mainTemplate); !ok {
		return nil, errors.New("Main template not found in model")
	}
	templates, err := extractTemplateFiles(m.Files)
	if err != nil {
		return nil, errors.Wrap(err, "Error with Model attachments")
	}
	tmpl2, err := extractTemplateFiles(m.Theme.Files)
	if err != nil {
		return nil, errors.Wrap(err, "Error with Theme attachments")
	}
	for k, v := range tmpl2 {
		templates[k] = v
	}

	// Rename to match the masterTemplate expectation
	templates["template.html"] = templates[mainTemplate]
	delete(templates, mainTemplate)
	// We pass face 0 here; the FuncMap will be replaced with the proper face before execution
	funcs, err := m.FuncMap()
	if err != nil {
		return nil, err
	}
	tmpl, err := template.New("template").Funcs(funcs).Parse(masterTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing master template")
	}
	for filename, t := range templates {
		tmplName := strings.TrimPrefix(filename, "!"+*m.Name+".")
		content := fmt.Sprintf("{{define \"%s\"}}%s{{end}}", tmplName, t)
		if _, err := tmpl.Parse(content); err != nil {
			return nil, errors.Wrapf(err, "Error parsing template file `%s`", filename)
		}
	}
	if css, ok := m.Theme.Files.GetFile("$main.css"); ok {
		if _, err := tmpl.Parse(fmt.Sprintf(`{{define "style.css"}}%s{{end}}`, css.Content)); err != nil {
			return nil, errors.Wrapf(err, "faild to parse $main.css")
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
		att, ok := v.GetFile(filename)
		if !ok {
			return nil, errors.Errorf("Error fetching expected file '%s' from Model", filename)
		}
		if _, ok := templateTypes[att.ContentType]; ok {
			templates[filename] = string(att.Content)
		}
	}
	return templates, nil
}

var masterTemplate = `
{{- $g := . -}}
<!DOCTYPE html>
<html>
<head>
	<title>FB Card</title>
	<base href="{{ .BaseURI }}">
	<meta charset="UTF-8">
	<link rel="stylesheet" type="text/css" href="css/cardframe.css">
<script type="text/javascript">
'use strict';
var FB = {
	face: {{ .Face }},
	card: {{ .Card }},
	note: {{ .Note }}
};
</script>
<script type="text/javascript" src="js/cardframe.js"></script>
<script type="text/javascript">{{ block "script.js" $g }}{{end}}</script>
<style>{{ block "style.css" $g }}{{end}}</style>
</style>
</head>
<body>{{ block "template.html" $g }}{{end}}</body>
</html>
`
