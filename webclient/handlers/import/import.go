package import_handler

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"time"

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
			return
		}
	}
	fmt.Printf("Done with import\n")
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
			modelMap, tmplMap, err := storeModels(collection)
			if err != nil {
				return err
			}
			noteMap, err := storeNotes(collection, modelMap)
			if err != nil {
				return err
			}
			deckMap, err := storeDecks(collection)
			if err != nil {
				return err
			}
			if err := storeCards(collection, deckMap, noteMap, tmplMap); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

type idmap map[int64]string
type tmplmap map[string][]string
type templ struct {
	Id   string
	Name string
}

var nameToIdRE = regexp.MustCompile("[[:space:]]")

func storeModels(c *anki.Collection) (idmap, tmplmap, error) {
	modelMap := make(idmap)
	templateMap := make(tmplmap)
	dbName := "user-" + util.CurrentUser()
	db := pouchdb.New(dbName)
	for _, m := range c.Models {
		if m.Type == anki.ModelTypeCloze {
			fmt.Printf("Cloze Models not yet supported\n")
			continue
		}
		modelUuid := m.AnkiId()
		model := data.Model{
			Id:          modelUuid,
			Name:        m.Name,
			Description: "Anki Model " + m.Name,
			Type:        "Model",
			Modified:    time.Now(),
			Created:     time.Now(),
			Comment:     "Imported from Anki on " + time.Now().String(),
		}
		modelMap[m.Id] = model.Id
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
		tmpls := make([]templ, len(m.Templates))
		templateMap[model.Id] = make([]string, len(m.Templates))
		for i, t := range m.Templates {
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
			tmpls[i] = templ{
				Id:   nameToIdRE.ReplaceAllString(t.Name, "_"),
				Name: t.Name,
			}
			templateMap[model.Id][i] = tmpls[i].Id
		}
		buf := new(bytes.Buffer)
		if err := masterTmpl.Execute(buf, tmpls); err != nil {
			return nil, nil, err
		}
		attachments = append(attachments, pouchdb.Attachment{
			Name: "template.html",
			Type: data.HTMLTemplateContentType,
			Body: buf,
		})
		rev, err := db.Put(model)
		if err != nil {
			return nil, nil, err
		}
		for _, a := range attachments {
			rev, err = db.PutAttachment(model.Id, &a, rev)
			if err != nil {
				return nil, nil, err
			}
		}
		if err != nil {
			return nil, nil, err
		}
	}
	return modelMap, templateMap, nil
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

func storeDecks(c *anki.Collection) (idmap, error) {
	deckMap := make(idmap)
	dbName := "user-" + util.CurrentUser()
	db := pouchdb.New(dbName)
	for _, d := range c.Decks {
		deckUuid := "deck-" + uuid.New()
		deck := data.Deck{
			Id:          deckUuid,
			Name:        d.Name,
			AnkiId:      d.AnkiId(),
			Description: d.Description,
			Type:        "deck",
			Modified:    time.Now(),
			Created:     time.Now(),
			Comment:     "Imported from Anki on " + time.Now().String(),
		}
		_, err := db.Put(deck)
		if err != nil {
			return nil, err
		}
		if err := storeDeckConfig(c, d.ConfigId, deckUuid); err != nil {
			return nil, err
		}
		deckMap[d.Id] = deck.Id
	}
	return deckMap, nil
}

func storeDeckConfig(c *anki.Collection, deckId int64, deckUuid string) error {
	dbName := "user-" + util.CurrentUser()
	db := pouchdb.New(dbName)
	for _, dc := range c.DeckConfig {
		if dc.Id == deckId {
			conf := data.DeckConfig{
				Id:              "deckconf-" + uuid.New(),
				Type:            "deckconf",
				DeckId:          deckUuid,
				Modified:        time.Now(),
				Created:         time.Now(),
				MaxDailyReviews: uint16(dc.Reviews.PerDay),
				MaxDailyNew:     uint16(dc.New.PerDay),
			}
			if _, err := db.Put(conf); err != nil {
				return err
			}
		}
		// There's only ever one DeckConfig per Deck, so return as soon as
		// we find it
		return nil
	}
	return nil
}

type noteNode struct {
	Uuid  string
	Model string
}
type notemap map[int64]noteNode

func storeNotes(c *anki.Collection, modelMap idmap) (notemap, error) {
	noteMap := make(notemap)
	dbName := "user-" + util.CurrentUser()
	db := pouchdb.New(dbName)
	for _, n := range c.Notes {
		modelUuid, ok := modelMap[n.ModelId]
		if !ok {
			// This isn't necessarily an error, since we skip cloze models, so just ignore for now
			continue
			// 			return nil, fmt.Errorf("Found note (id=%d) with no model", n.Id)
		}
		note := data.Note{
			Id:          "note-" + uuid.New(),
			AnkiId:      n.AnkiId(),
			Type:        "note",
			ModelId:     modelUuid,
			Modified:    time.Now(),
			Created:     time.Now(),
			Tags:        n.Tags,
			FieldValues: n.Fields,
		}
		if _, err := db.Put(note); err != nil {
			return nil, err
		}
		noteMap[n.Id] = noteNode{Uuid: note.Id, Model: modelUuid}
	}
	return noteMap, nil
}

func storeCards(c *anki.Collection, deckMap idmap, noteMap notemap, tmplMap tmplmap) error {
	related := make(map[string][]string)
	var cards []data.Card
	for _, c := range c.Cards {
		var note noteNode
		var deckUuid string
		var ok bool
		if deckUuid, ok = deckMap[c.DeckId]; !ok {
			return errors.New("Found card that doesn't belong to a deck")
		}
		if note, ok = noteMap[c.NoteId]; !ok {
			// Probably due to cloze
			continue
			// 			return fmt.Errorf("Found card (%d) with no note", c.Id)
		}
		card := data.Card{
			Id:         "card-" + uuid.New(),
			AnkiId:     c.AnkiId(),
			Type:       "card",
			NoteId:     note.Uuid,
			DeckId:     deckUuid,
			Modified:   time.Now(),
			Created:    time.Now(),
			Reviews:    c.Reps,
			Lapses:     c.Lapses,
			Interval:   c.Interval,
			SRSFactor:  float32(c.Factor) / 1000,
			TemplateId: note.Model + "-" + tmplMap[note.Model][c.Ord],
		}
		switch c.Type {
		case anki.CardTypeLearning:
			card.Due = time.Unix(0, 0).AddDate(0, 0, int(c.Due))
		case anki.CardTypeDue:
			card.Due = time.Unix(c.Due, 0)
		}
		if c.Queue == anki.QueueTypeSuspended {
			card.Suspended = true
		}
		if rel, ok := related[card.NoteId]; ok {
			rel = append(rel, card.Id)
		} else {
			related[card.NoteId] = []string{card.Id}
		}
		cards = append(cards, card)
	}
	dbName := "user-" + util.CurrentUser()
	db := pouchdb.New(dbName)
	for _, card := range cards {
		rel := related[card.NoteId]
		card.RelatedCards = make([]string, len(rel)-1)
		var i int
		for _, r := range rel {
			fmt.Printf("Comparing %s to %s\n", r, card.Id)
			if r != card.Id {
				card.RelatedCards[i] = r
				i++
			}
		}
		if _, err := db.Put(card); err != nil {
			return err
		}
	}
	return nil
}

type Card struct {
	Suspeended   bool     `json:"Suspended"`
	RelatedCards []string `json:"Related"`
}

/*
type Collection struct {
	Created        time.Time
	Modified       time.Time
	SchemaModified time.Time
	Ver            int
	LastSync       time.Time
	Config         Config
	Cards          []*Card
	Revlog         []*Review
}
*/
