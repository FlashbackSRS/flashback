package model

import (
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/flimzy/log"
	"github.com/pkg/errors"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/controllers"
)

// // Model is an interface to a model.
// type Model interface {
// 	Template() (*template.Template, error)
// }

// fbModel is a wrapper around a *fb.Model
type fbModel struct {
	*fb.Model
}

// var _ Model = &fbModel{}

const mainCSS = "$main.css"

func (m *fbModel) Template() (*template.Template, error) {
	start := time.Now()
	log.Debug("started Template()")
	defer func() {
		finish := time.Now()
		log.Debug("finished Template (%v)", finish.Sub(start))
	}()
	mc, err := controllers.GetModelController(m.Type)
	if err != nil {
		return nil, err
	}
	mainTemplate := fmt.Sprintf("$template.%d.html", m.ID)
	if _, ok := m.Files.GetFile(mainTemplate); !ok {
		return nil, errors.New("Main template not found in model")
	}
	templates := extractTemplateFiles(m.Files)
	tmpl2 := extractTemplateFiles(m.Theme.Files)
	for k, v := range tmpl2 {
		templates[k] = v
	}

	// Rename to match the masterTemplate expectation
	templates["template.html"] = templates[mainTemplate]
	delete(templates, mainTemplate)
	delete(templates, mainCSS)

	var funcs template.FuncMap
	if funcMapper, ok := mc.(controllers.FuncMapper); ok {
		funcs = funcMapper.FuncMap(nil, 0)
	}
	tmpl, err := template.New("template").Funcs(funcs).Parse(masterTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "Error parsing master template") // This should never happen
	}
	for filename, t := range templates {
		tmplName := strings.TrimPrefix(filename, "!"+m.Name+".")
		content := fmt.Sprintf("{{define \"%s\"}}%s{{end}}", tmplName, t)
		if _, err := tmpl.Parse(content); err != nil {
			return nil, errors.Wrapf(err, "Error parsing template file `%s`", filename)
		}
	}
	if css, ok := m.Theme.Files.GetFile(mainCSS); ok {
		if _, err := tmpl.Parse(fmt.Sprintf(`{{define "style.css"}}%s{{end}}`, css.Content)); err != nil {
			return nil, errors.Wrapf(err, "failed to parse "+mainCSS)
		}
	}
	return tmpl, nil
}

var templateTypes = map[string]struct{}{
	fb.TemplateContentType: struct{}{},
	"text/css":             struct{}{},
	"script/javascript":    struct{}{},
}

func extractTemplateFiles(v *fb.FileCollectionView) map[string]string {
	templates := make(map[string]string)
	for _, filename := range v.FileList() {
		att, _ := v.GetFile(filename)
		if _, ok := templateTypes[att.ContentType]; ok {
			templates[filename] = string(att.Content)
		}
	}
	return templates
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

func (m *fbModel) FuncMap(card *fbCard, face int) (template.FuncMap, error) {
	mc, err := controllers.GetModelController(m.Type)
	if err != nil {
		return nil, err
	}
	if funcMapper, ok := mc.(controllers.FuncMapper); ok {
		// card may be nil during template creation
		var c *fb.Card
		if card != nil {
			c = card.Card
		}
		return funcMapper.FuncMap(c, face), nil
	}
	return nil, nil
}

func (m *fbModel) IframeScript() ([]byte, error) {
	mc, err := controllers.GetModelController(m.Type)
	if err != nil {
		return nil, err
	}
	return mc.IframeScript(), nil
}
