package study_handler

import (
	// 	"bytes"
	// 	"errors"
	"fmt"
	// 	"html/template"
	"net/url"
	// 	"strings"

	// 	"github.com/pborman/uuid"

	// 	"github.com/flimzy/go-pouchdb"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/gopherjs/jsbuiltin"
	// 	"golang.org/x/net/html"

	"github.com/flimzy/flashback-model"
	// 	"github.com/flimzy/flashback/util"
	"github.com/flimzy/flashback/repository"
)

var jQuery = jquery.NewJQuery

func BeforeTransition(event *jquery.Event, ui *js.Object, p url.Values) bool {
	u, err := repo.CurrentUser()
	if err != nil {
		fmt.Printf("No user logged in: %s\n", err)
		return false
	}
	fmt.Printf("card = %s\n", p.Get("card"))
	go func() {
		container := jQuery(":mobile-pagecontainer")
		// Ensure the indexes are created before trying to use them
		u.DB()

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

func getCard() (*fb.Card, error) {
	// 	u, err := repo.CurrentUser()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	db, err := repo.UserDB(u)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	doc := make(map[string][]fb.Card)
	// 	query := map[string]interface{}{
	// 		"selector": map[string]string{"type": "card"},
	// 		"limit":    1,
	// 	}
	// 	if err := db.Find(query, &doc); err != nil {
	// 		return nil, err
	// 	}
	// 	card := doc["docs"][0]
	// 	return &card, nil
	return nil, nil
}

/*
func getModel(id string) (*fb.Model, error) {
	db, err := util.UserDb()
	if err != nil {
		return nil, err
	}
	var model fb.Model
	err = db.Get(id, &model, pouchdb.Options{})
	return &model, err
}

func getNote(id string) (*fb.Note, error) {
	db, err := util.UserDb()
	if err != nil {
		return nil, err
	}
	var note fb.Note
	err = db.Get(id, &note, pouchdb.Options{})
	return &note, err
}
*/

func getCardBodies(card *fb.Card) (string, string, error) {
	return "", "", nil
}

/*
	note, err := getNote("") // FIXME: this isn't right; and this logic should move to repository.go
	if err != nil {
		return "", "", err
	}
	model, err := getModel("") // FIXME: this isn't right, and this logic should move to repository.go
	if err != nil {
		return "", "", err
	}

	db, err := util.UserDb()
	if err != nil {
		return "", "", err
	}

	templates := make(map[string]string)
	for filename, a := range model.Attachments {
		if a.Type == data.HTMLTemplateContentType || a.Type == "text/css" || a.Type == "script/javascript" {
			att, err := db.Attachment(model.ID, filename, model.Rev)
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
		content := fmt.Sprintf("{{define \"%s\"}}%s{{end}}", filename, t)
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

	htmlDoc := new(bytes.Buffer)
	if err := tmpl.Execute(htmlDoc, ctx); err != nil {
		return "", "", err
	}
	// fmt.Printf("Body: '%s'\n", base64.StdEncoding.EncodeToString( body.Bytes() ))
	doc, err := html.Parse(bytes.NewReader(htmlDoc.Bytes()))
	if err != nil {
		return "", "", err
	}
	tmp := strings.Split(card.TemplateId, "/")
	fmt.Printf("Target id: %s, target class: %s\n", tmp[1], "question")
	body := findBody(doc)
	if body == nil {
		return "", "", errors.New("No <body> in template output")
	}

	container := findContainer(body.FirstChild, tmp[1], "question")
	if container == nil {
		fmt.Printf("No matching div found in template output\n")
		return htmlDoc.String(), ctx.IframeId, nil
	}
	fmt.Printf("Deleting divs\n")
	for c := body.FirstChild; c != nil; c = body.FirstChild {
		body.RemoveChild(c)
	}
	// 	container.RemoveChild(inner)
	fmt.Printf("Appending target div's inner conent\n")
	// 	for c := container.FirstChild; c != nil; c = container.FirstChild {
	// 		container.RemoveChild( c )
	// 		body.AppendChild( c )
	// 	}
	inner := container.FirstChild
	inner.Parent = body
	body.FirstChild = inner

	newBody := new(bytes.Buffer)
	if err := html.Render(newBody, doc); err != nil {
		return "", "", err
	}
	nbString := newBody.String()
	fmt.Printf("original size = %d\n", len(htmlDoc.String()))
	fmt.Printf("new body size = %d\n", len(nbString))
	return nbString, ctx.IframeId, nil
}

func findBody(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "body" {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if body := findBody(c); body != nil {
			return body
		}
	}
	return nil
}

func findContainer(n *html.Node, targetId, targetClass string) *html.Node {
	if n == nil {
		return nil
	}
	if n.Type == html.ElementNode && n.Data == "div" {
		var class, id string
		for _, a := range n.Attr {
			switch a.Key {
			case "class":
				class = a.Val
			case "data-id":
				id = a.Val
			}
			if class != "" && id != "" {
				break
			}
		}
		if class == targetClass && id == targetId {
			return n
		}
	}
	return findContainer(n.NextSibling, targetId, targetClass)
}

type cardContext struct {
	IframeId string
	Card     *fb.Card
	Note     *fb.Note
	Model    *fb.Model
	Deck     *fb.Deck
	BaseURI  string
	Fields   map[string]template.HTML
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
	iframeId: '{{ .IframeId }}',
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
*/
