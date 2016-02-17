package import_handler

import (
	"archive/zip"
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/flimzy/flashback/anki"
	"github.com/flimzy/flashback/data"
	"github.com/flimzy/flashback/util"
	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/web/file"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/pborman/uuid"
	"honnef.co/go/js/console"
)

var jQuery = jquery.NewJQuery

func BeforeTransition(event *jquery.Event, ui *js.Object) bool {

	go func() {
		container := jQuery(":mobile-pagecontainer")
		jQuery("#importnow", container).On("click", func() {
			fmt.Printf("Attempting to import something...\n")
			go DoImport()
		})
		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()
	}()

	return true
}

func DoImport() {
	files := jQuery("#apkg", ":mobile-pagecontainer").Get(0).Get("files")
	for i := 0; i < files.Length(); i++ {
		err := importFile(file.Internalize(files.Index(i)))
		if err != nil {
			fmt.Printf("Error importing file: %s\n", err)
		}
	}
}

func importFile(f file.File) error {
	fmt.Printf("Gonna pretend to import %s now\n", f.Name())
	z, err := zip.NewReader(bytes.NewReader(f.Bytes()), f.Size())
	if err != nil {
		return err
	}
	for _, file := range z.File {
		fmt.Printf("Archive contains %s\n", file.FileHeader.Name)
		if file.FileHeader.Name == "collection.anki2" {
			// Found the SQLite database
			rc, err := file.Open()
			if err != nil {
				return err
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(rc)
			collection, err := readSQLite(buf.Bytes())
			if err != nil {
				return err
			}
			console.Log(collection)
			if err := storeModels(collection); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func storeModels(c *anki.Collection) error {
	dbName := "user-" + util.CurrentUser()
	db := pouchdb.New(dbName)
	for _, m := range c.Models {
		if m.Type == anki.ModelTypeCloze {
			fmt.Printf("Cloze Models not yet supported\n")
			continue
		}
		model := data.Model{
			Id:          "model-" + uuid.New(),
			Name:        m.Name,
			Description: "Anki Model " + m.Name,
			Type:        "Model",
		}
		for _, f := range m.Fields {
			model.Fields = append(model.Fields, &data.Field{
				Name: f.Name,
			})
		}
		attachments := []pouchdb.Attachment{
			pouchdb.Attachment{
				Name: "style.css",
				Type: "text/css",
				Body: strings.NewReader(m.CSS),
			},
		}
		var tmpls []struct {
			Id   string
			Name string
		}
		nameToIdRE := regexp.MustCompile("[[:space:]]")
		for _, t := range m.Templates {
			attachments = append(attachments, pouchdb.Attachment{
				Name: t.Name + " front.html",
				Type: data.HTMLTemplateContentType,
				Body: strings.NewReader(t.QuestionFormat),
			})
			attachments = append(attachments, pouchdb.Attachment{
				Name: t.Name + " back.html",
				Type: data.HTMLTemplateContentType,
				Body: strings.NewReader(t.AnswerFormat),
			})
			tmpls = append(tmpls, struct {
				Id   string
				Name string
			}{
				Id:   nameToIdRE.ReplaceAllString(t.Name, "_"),
				Name: t.Name,
			})
		}
		buf := new(bytes.Buffer)
		if err := masterTmpl.Execute(buf, tmpls); err != nil {
			return err
		}
		attachments = append(attachments, pouchdb.Attachment{
			Name: "template.html",
			Type: data.HTMLTemplateContentType,
			Body: buf,
		})
		rev, err := db.Put(model)
		if err != nil {
			return err
		}
		for _, a := range attachments {
			fmt.Printf("Before rev = %s\n", rev)
			rev, err = db.PutAttachment(model.Id, &a, rev)
			fmt.Printf("After rev = %s\n", rev)
			if err != nil {
				return err
			}
		}
		fmt.Printf("Put model #%s\n", rev)
	}
	return nil
}

var masterTmpl = template.Must(template.New("template.html").Delims("[[", "]]").Parse(`
<html>
<head>
<style>
{{ template "style.css" }}
</style>
<script>
{{ template "script.js" }}
</script>
</head>
<body>
{{ $g := . }}
[[ range . ]]
	<div id="front-{{ .Id }}">
		{{template "[[ .Name ]] front.html" $g}}
	</div>
	<div id="back-{{ . }}">
		{{template "[[ .Name ]] back.html" $g}}
	</div>
[[ end ]]
</body>
</html>
`))
