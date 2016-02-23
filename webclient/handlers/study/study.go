package study_handler

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"

	"github.com/pborman/uuid"

	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/go-pouchdb/plugins/find"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/gopherjs/jsbuiltin"
	// 	"golang.org/x/net/html"

	"github.com/flimzy/flashback/data"
	"github.com/flimzy/flashback/util"
	// 	"honnef.co/go/js/console"
)

var jQuery = jquery.NewJQuery

func BeforeTransition(event *jquery.Event, ui *js.Object) bool {
	go func() {
		container := jQuery(":mobile-pagecontainer")
		// Ensure the indexes are created before trying to use them
		<-util.InitUserDb()

		card, err := getCard()
		if err != nil {
			fmt.Printf("Error fetching card: %s\n", err)
			return
		}
		body, iframeId, err := getCardBodies(card)
		if err != nil {
			fmt.Printf("Error reading card: %s\n", err)
		}

		iframe := js.Global.Get("document").Call("createElement", "iframe")
		iframe.Call("setAttribute", "sandbox", "allow-scripts")
		iframe.Call("setAttribute", "seamless", nil)
		iframe.Set("id", iframeId)
		iframe.Set("src", "data:text/html;charset=utf-8,"+jsbuiltin.EncodeURI(body))

		js.Global.Get("document").Call("getElementById", "cardframe").Call("appendChild", iframe)

		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()
	}()

	return true
}

func getCard() (*data.Card, error) {
	dbfind := find.New(util.UserDb())
	doc := make(map[string][]data.Card)
	err := dbfind.Find(map[string]interface{}{
		"selector": map[string]string{"$type": "card"},
		"limit":    1,
	}, &doc)
	if err != nil {
		return nil, err
	}
	card := doc["docs"][0]
	return &card, nil
}

func getModel(id string) (*data.Model, error) {
	db := util.UserDb()
	var model data.Model
	err := db.Get(id, &model, pouchdb.Options{})
	return &model, err
}

func getNote(id string) (*data.Note, error) {
	db := util.UserDb()
	var note data.Note
	err := db.Get(id, &note, pouchdb.Options{})
	return &note, err
}

func getCardBodies(card *data.Card) (string, string, error) {
	note, err := getNote(card.NoteId)
	if err != nil {
		return "", "", err
	}
	model, err := getModel(note.ModelId)
	if err != nil {
		return "", "", err
	}

	db := util.UserDb()

	templates := make(map[string]string)
	for filename, a := range model.Attachments {
		if a.Type == data.HTMLTemplateContentType || a.Type == "text/css" || a.Type == "script/javascript" {
			att, err := db.Attachment(model.Id, filename, model.Rev)
			if err != nil {
				return "", "", err
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(att.Body)
			templates[filename] = buf.String()
		}
	}
	if _, ok := templates["template.html"]; !ok {
		return "", "", errors.New("No master template defined")
	}
	tmpl, err := template.New("template").Parse(masterTemplate)
	if err != nil {
		return "", "", err
	}
	for filename, t := range templates {
		content := fmt.Sprintf(`
{{define "%s"}}
%s
{{end}}
		`, filename, t)
		if _, err := tmpl.Parse(content); err != nil {
			return "", "", err
		}
	}

	ctx := cardContext{
		IframeId: uuid.New(),
		Card:     card,
		Note:     note,
		Model:    model,
		BaseURI:  util.BaseURI(),
		Fields:   make(map[string]template.HTML),
	}

	for i, f := range model.Fields {
		ctx.Fields[f.Name] = template.HTML(note.FieldValues[i])
	}

	body := new(bytes.Buffer)
	if err := tmpl.Execute(body, ctx); err != nil {
		return "", "", err
	}

	return body.String(), ctx.IframeId, nil
}

type cardContext struct {
	IframeId string
	Card     *data.Card
	Note     *data.Note
	Model    *data.Model
	Deck     *data.Deck
	BaseURI  string
	Fields   map[string]template.HTML
}

var masterTemplate = `
<html>
<head>
	<base href="{{ .BaseURI }}">
	<meta http-equiv="Content-Security-Policy"
		content="
		script-src 'unsafe-inline' {{ .BaseURI }}
	">
<script type="text/javascript">
'use strict';
var FB = {
	iframeId: '{{ .IframeId }}',
	card: {{ .Card }},
	note: {{ .Note }},
	model: {{ .Model }},
};
</script>
<script type="text/javascript" src="js/cardframe.js"></script>
<script type="text/javascript">
{{ block "script.js" .Fields }}{{end}}
</script>
<link rel="stylesheet" type="text/css" href="style.css">
</head>
<body>
{{ block "template.html" .Fields }}{{end}}
</body>
</html>
`
