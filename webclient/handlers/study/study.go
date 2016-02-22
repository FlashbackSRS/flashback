package study_handler

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"
	"github.com/gopherjs/jquery"
	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/go-pouchdb/plugins/find"
// 	"golang.org/x/net/html"

	"github.com/flimzy/flashback/util"
	"github.com/flimzy/flashback/data"
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
		body, err := getCardBodies(card)
		if err != nil {
			fmt.Printf("Error reading card: %s\n", err)
		}
		
		iframe := js.Global.Get("document").Call("getElementById", "cardframe").Call("cloneNode")
// 		sandbox := iframe.Call("getAttribute", "sandbox")
		iframe.Call("removeAttribute", "sandbox")
		
//		iframe := js.Global.Get("document").Call("createElement", "iframe")
		iframe.Set("src", "data:text/html;charset=utf-8," + jsbuiltin.EncodeURI(body))
//		iframe.Call("setAttribute", "sandbox", sandbox)

		jQuery("#cardframe", container).ReplaceWith(iframe)
		
		doc := iframe.Get("contentWindow").Get("document")
		doc.Call("open")
		doc.Call("write", body)
		doc.Call("close")

		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()
	}()

	return true
}

func getCard() (*data.Card,error) {
	dbfind := find.New( util.UserDb() )
	doc := make(map[string][]data.Card)
	err := dbfind.Find(map[string]interface{}{
		"selector": map[string]string{ "$type": "card" },
		"limit": 1,
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

func getCardBodies(card *data.Card) (string, error) {
	note, err := getNote( card.NoteId )
	if err != nil {
		return "", err
	}
	model, err := getModel( note.ModelId )
	if err != nil {
		return "", err
	}
	
	db := util.UserDb()
	
	templates := make(map[string]string)
	for filename, a := range model.Attachments {
		if a.Type == data.HTMLTemplateContentType || a.Type == "text/css" || a.Type == "script/javascript" {
			att, err := db.Attachment(model.Id, filename, model.Rev)
			if err != nil {
				return "", err
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(att.Body)
			templates[filename] = buf.String()
		}
	}
	if _, ok := templates["template.html"]; ! ok {
		return "", errors.New("No master template defined")
	}
	tmpl, err := template.New("template").Parse(masterTemplate)
	if err != nil {
		return "", err
	}
	for filename, t := range templates {
		content := fmt.Sprintf(`
{{define "%s"}}
%s
{{end}}
		`, filename, t)
		if _, err := tmpl.Parse(content); err != nil {
			return "", err
		}
	}
	
	content := make(map[string]string)
	for i,f := range model.Fields {
		content[ f.Name ] = note.FieldValues[i]
	}
	
	body := new(bytes.Buffer)
	if err := tmpl.Execute(body, content); err != nil {
		return "", err
	}

	return body.String(), nil
}

var masterTemplate = `
<html>
<head>
    <meta http-equiv="Content-Security-Policy"
        content="
        script-src 'self' 'unsafe-inline'
            ">
<script type="text/javascript" src="js/cardframe.js"></script>
<style>
{{ block "style.css" . }}{{end}}
</style>
<script>
{{ block "script.js" . }}{{end}}
</script>
</head>
<body>
{{ block "template.html" . }}{{end}}
</body>
</html>
`
