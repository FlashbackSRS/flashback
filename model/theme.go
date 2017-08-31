package model

import (
	"context"
	"fmt"
	"html/template"
	"strings"

	"github.com/flimzy/log"
	"github.com/pkg/errors"

	fb "github.com/FlashbackSRS/flashback-model"
)

// fbModel is a wrapper around a *fb.Model
type fbModel struct {
	*fb.Model
	db attachmentGetter
}

func (m *fbModel) getCachedAttachment(filename string) (*fb.Attachment, error) {
	att, ok := m.Files.GetFile(filename)
	if ok {
		return att, nil
	}
	att, ok = m.Theme.Files.GetFile(filename)
	if ok {
		return att, nil
	}
	return nil, fmt.Errorf("attachment '%s' not found", filename)
}

func (m *fbModel) getAttachment(ctx context.Context, filename string) (*fb.Attachment, error) {
	att, err := m.getCachedAttachment(filename)
	if err != nil {
		return nil, err
	}
	if len(att.Content) != 0 {
		return att, nil
	}
	dbAtt, err := getAttachment(ctx, m.db, m.Model.Theme.ID, filename)
	if err != nil {
		return nil, err
	}
	m.Files.SetFile(filename, att.ContentType, dbAtt.Content)
	att, _ = m.Files.GetFile(filename)
	return att, nil
}

type templateCacheItem struct {
	rev  string
	tmpl *template.Template
}

type templateCache map[string]templateCacheItem

func templateCacheKey(m *fb.Model) string {
	return fmt.Sprintf("%s %d", m.Theme.ID, m.ID)
}

func (c templateCache) Get(m *fb.Model) (*template.Template, bool) {
	key := templateCacheKey(m)
	tmpl, ok := c[key]
	if !ok || tmpl.rev != m.Theme.Rev {
		return nil, false
	}
	return tmpl.tmpl, true
}

func (c templateCache) Set(m *fb.Model, tmpl *template.Template) {
	key := templateCacheKey(m)
	c[key] = templateCacheItem{
		rev:  m.Theme.Rev,
		tmpl: tmpl,
	}
}

var globalTemplateCache = make(templateCache)

const mainCSS = "$main.css"

func (m *fbModel) Template(ctx context.Context) (*template.Template, error) {
	defer profile("Template")()
	if tmpl, cached := globalTemplateCache.Get(m.Model); cached {
		log.Debugf("Returning cached template")
		return tmpl, nil
	}
	tmpl, err := m.template(ctx)
	if err != nil {
		return nil, err
	}
	globalTemplateCache.Set(m.Model, tmpl)
	return tmpl, nil
}

func (m *fbModel) template(ctx context.Context) (*template.Template, error) {
	defer profile("template")()
	mc, err := GetModelController(m.Type)
	if err != nil {
		return nil, err
	}
	mainTemplate := fmt.Sprintf("$template.%d.html", m.ID)
	if _, ok := m.Files.GetFile(mainTemplate); !ok {
		return nil, fmt.Errorf("main template '%s' not found in model", mainTemplate)
	}

	templates, err := m.extractTemplateFiles(ctx)
	if err != nil {
		return nil, err
	}

	// Rename to match the masterTemplate expectation
	templates["template.html"] = templates[mainTemplate]
	templates["style.css"] = templates[mainCSS]
	delete(templates, mainTemplate)
	delete(templates, mainCSS)

	var funcs template.FuncMap
	if funcMapper, ok := mc.(FuncMapper); ok {
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
	return tmpl, nil
}

var templateTypes = map[string]struct{}{
	fb.TemplateContentType: struct{}{},
	"text/css":             struct{}{},
	"script/javascript":    struct{}{},
}

func (m *fbModel) extractTemplateFiles(ctx context.Context) (map[string]string, error) {
	templates := make(map[string]string)
	for _, filename := range m.Files.FileList() {
		att, err := m.getAttachment(ctx, filename)
		if err != nil {
			return nil, err
		}
		templates[filename] = string(att.Content)
	}
	for _, filename := range m.Theme.Files.FileList() {
		att, err := m.getAttachment(ctx, filename)
		if err != nil {
			return nil, err
		}
		templates[filename] = string(att.Content)
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
</head>
<body>{{ block "template.html" $g }}{{end}}</body>
</html>
`

func (m *fbModel) FuncMap(card *Card, face int) (template.FuncMap, error) {
	mc, err := GetModelController(m.Type)
	if err != nil {
		return nil, err
	}
	if funcMapper, ok := mc.(FuncMapper); ok {
		return funcMapper.FuncMap(card, face), nil
	}
	return nil, nil
}

func (m *fbModel) IframeScript() ([]byte, error) {
	mc, err := GetModelController(m.Type)
	if err != nil {
		return nil, err
	}
	return mc.IframeScript(), nil
}
