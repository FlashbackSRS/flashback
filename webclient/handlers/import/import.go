package import_handler

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"path/filepath"
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
	zipMap := make(map[string]*zip.File)
	for _, file := range z.File {
		zipMap[ file.FileHeader.Name ] = file
	}
	mediaMap, err := extractMediaMap(zipMap)
	if err != nil {
		return err
	}
	collection, err := extractCollection(zipMap)
	if err != nil {
		return err
	}
	console.Log(collection)
	modelMap, err := storeModels(collection)
	if err != nil {
		return err
	}
	noteMap, err := storeNotes(collection, modelMap, mediaMap)
	if err != nil {
		return err
	}
	deckMap, err := storeDecks(collection)
	if err != nil {
		return err
	}
	if err := storeCards(collection, deckMap, noteMap); err != nil {
		return err
	}
	return nil
}

func extractMediaMap(z map[string]*zip.File) (map[string]*zip.File, error) {
	file, ok := z["media"]
	if ! ok {
		return nil, errors.New("Did not find 'media'. Invalid Anki package")
	}
	var fileMap map[string]string
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(rc)
	if err := json.Unmarshal(buf.Bytes(), &fileMap); err != nil {
		return nil, err
	}
	mediaMap := make(map[string]*zip.File)
	for archiveName,fileName := range fileMap {
		mediaMap[fileName] = z[archiveName]
	}
	return mediaMap, nil
}

func extractCollection(z map[string]*zip.File) (*anki.Collection, error) {
	file, ok := z["collection.anki2"]
	if ! ok {
		return nil, errors.New("Did not find 'collection.anki2'. Invalid Anki package")
	}
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(rc)
	return readSQLite(buf.Bytes())
}

type idmap map[int64]string

var nameToIdRE = regexp.MustCompile("[[:space:]]")

func storeModels(c *anki.Collection) (idmap, error) {
	modelMap := make(idmap)
	dbName := "user-" + util.CurrentUser()
	db := pouchdb.New(dbName)
	for _, m := range c.Models {
		if m.Type == anki.ModelTypeCloze {
			fmt.Printf("Cloze Models not yet supported\n")
			continue
		}
		modelUuid := m.AnkiId()
		modelMap[m.Id] = modelUuid
		var model data.Model
		// Check for duplicates
		if err := db.Get(modelUuid, &model, pouchdb.Options{}); err == nil {
			continue
		}
		model = data.Model{
			Id:          modelUuid,
			Name:        m.Name,
			Description: "Anki Model " + m.Name,
			Type:        "Model",
			Modified:    time.Now(),
			Created:     time.Now(),
			Comment:     "Imported from Anki on " + time.Now().String(),
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
		tmpls := make([]string, len(m.Templates))
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
			tmpls[i] = t.Name
		}
		buf := new(bytes.Buffer)
		if err := masterTmpl.Execute(buf, tmpls); err != nil {
			return nil, err
		}
		attachments = append(attachments, pouchdb.Attachment{
			Name: "template.html",
			Type: data.HTMLTemplateContentType,
			Body: buf,
		})
		rev, err := db.Put(model)
		if err != nil {
			return nil, err
		}
		for _, a := range attachments {
			rev, err = db.PutAttachment(model.Id, &a, rev)
			if err != nil {
				return nil, err
			}
		}
		if err != nil {
			return nil, err
		}
	}
	return modelMap, nil
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
[[ range $i, $Name := . ]]
	<div id="front-[[ $i ]]">
		{{template "[[ $Name ]] front.html" $g}}
	</div>
	<div id="back-[[ $i ]]">
		{{template "[[ $Name ]] back.html" $g}}
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
		deckUuid := d.AnkiId()
		deckMap[d.Id] = deckUuid
		var deck data.Deck
		if err := db.Get(deckUuid, &deck, pouchdb.Options{}); err == nil {
			continue
		}
		deck = data.Deck{
			Id:          deckUuid,
			Name:        d.Name,
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

var soundRe = regexp.MustCompile("\\[sound:(.*?)\\]")
var imageRe = regexp.MustCompile("<img src=\"(.*?)\" />")

func storeNotes(c *anki.Collection, modelMap idmap, mediaMap map[string]*zip.File) (notemap, error) {
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
		noteUuid := n.AnkiId()
		noteMap[n.Id] = noteNode{Uuid: noteUuid, Model: modelUuid}
		var note data.Note
		if err := db.Get(noteUuid, &note, pouchdb.Options{}); err == nil {
			continue
		}
		note = data.Note{
			Id:          noteUuid,
			Type:        "note",
			ModelId:     modelUuid,
			Modified:    time.Now(),
			Created:     time.Now(),
			Tags:        n.Tags,
			FieldValues: n.Fields,
		}
		rev, err := db.Put(note)
		if err != nil {
			return nil, err
		}
		files := make(map[string][]string)
		files["audio"] = make([]string,0)
		files["image"] = make([]string,0)
		for _, field := range note.FieldValues {
			for _,match := range soundRe.FindAllStringSubmatch(field,-1) {
				files["audio"] = append( files["audio"], match[1])
			}
			for _, match := range imageRe.FindAllStringSubmatch(field,-1) {
				files["image"] = append( files["image"], match[1])
			}
		}
		for ftype, filenames := range files {
			for _, file := range filenames {
				contentType, ok := contentTypeMap[ftype][ filepath.Ext(file) ]
				if ! ok {
					fmt.Printf("Unknown content type for file '%s'\n", file)
					contentType = ftype + "/unknown"
				}

				rc, err := mediaMap[file].Open()
				if err != nil {
					return nil, err
				}
				defer rc.Close()
				attachment := pouchdb.Attachment{
					Name: file,
					Type: contentType,
					Body: rc,
				}
				rev, err = db.PutAttachment(note.Id, &attachment, rev)
			}
		}
	}
	return noteMap, nil
}

var contentTypeMap = map[string]map[string]string{
	"audio": map[string]string{
		".3gp": "audio/3gpp",
		".ogg": "audio/ogg",
		".mp3": "audio/mpeg",
		".spx": "audio/ogg",
		".wav": "audio/x-wav",
		".flac": "audio/flac",
	},
	"image": map[string]string{
		".jpg": "image/jpeg",
		".jpeg": "image/jpeg",
		".png": "image/png",
		".gif": "image/gif",
	},
}

func storeCards(c *anki.Collection, deckMap idmap, noteMap notemap) error {
	related := make(map[string]*[]string)
	var cards []data.Card
	dbName := "user-" + util.CurrentUser()
	db := pouchdb.New(dbName)
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
		cardUuid := c.AnkiId(note.Uuid)
		if rel, ok := related[note.Uuid]; ok {
			*rel = append(*rel, cardUuid)
		} else {
			related[note.Uuid] = &([]string{cardUuid})
		}
		var card data.Card
		if err := db.Get(cardUuid, &card, pouchdb.Options{}); err == nil {
			continue
		}
		card = data.Card{
			Id:         cardUuid,
			Type:       "card",
			NoteId:     note.Uuid,
			DeckId:     deckUuid,
			Modified:   time.Now(),
			Created:    time.Now(),
			Reviews:    c.Reps,
			Lapses:     c.Lapses,
			Interval:   c.Interval,
			SRSFactor:  float32(c.Factor) / 1000,
			TemplateId: fmt.Sprintf("%s/%d", note.Model, c.Ord),
			Suspended:  c.Queue == anki.QueueTypeSuspended,
		}
		switch c.Type {
			case anki.CardTypeLearning:
				card.Due = time.Unix(0, 0).AddDate(0, 0, int(c.Due))
			case anki.CardTypeDue:
				card.Due = time.Unix(c.Due, 0)
		}
		cards = append(cards, card)
	}
	for _, card := range cards {
		rel := related[card.NoteId]
		card.RelatedCards = make([]string, len(*rel)-1)
		var i int
		for _, r := range *rel {
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
