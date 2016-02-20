package import_handler

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"path/filepath"
	"reflect"
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
		zipMap[file.FileHeader.Name] = file
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
	if !ok {
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
	for archiveName, fileName := range fileMap {
		mediaMap[fileName] = z[archiveName]
	}
	return mediaMap, nil
}

func extractCollection(z map[string]*zip.File) (*anki.Collection, error) {
	file, ok := z["collection.anki2"]
	if !ok {
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
			if model.Modified.After(model.AnkiImported) {
				fmt.Printf("Model %d / %s has local changes, skipping.\n", m.Id, model.Id)
				continue
			}
			fmt.Printf("Gonna update model %d / %s\n", m.Id, model.Id)
			if err := updateModel(&model, m); err != nil {
				return nil, err
			}
			continue
		}
		model = data.Model{
			Id:           modelUuid,
			Rev:          model.Rev, // Preserve the Rev, if there is one, so the put succeeds
			Name:         m.Name,
			Description:  "Anki Model " + m.Name,
			Type:         "model",
			Modified:     m.Modified,
			AnkiImported: time.Now(),
			Comment:      "Imported from Anki on " + time.Now().String(),
		}
		for _, f := range m.Fields {
			model.Fields = append(model.Fields, &data.Field{
				Name: f.Name,
			})
		}
		attachments, err := modelAttachments(m)
		if err != nil {
			return nil, err
		}
		rev, err := db.Put(model)
		if err != nil {
			return nil, err
		}
		for _, a := range *attachments {
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

func modelAttachments(m *anki.Model) (*[]pouchdb.Attachment, error) {
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

	return &attachments, nil
}

func updateModel(model *data.Model, m *anki.Model) error {
	var changed bool
	if m.Name != model.Name {
		model.Name = m.Name
		model.Description = "Anki Model " + m.Name
		changed = true
	}
	var fields []*data.Field
	for _, f := range m.Fields {
		fields = append(fields, &data.Field{
			Name: f.Name,
		})
	}
	if !reflect.DeepEqual(fields, model.Fields) {
		model.Fields = fields
		changed = true
	}

	attachments, err := modelAttachments(m)
	if err != nil {
		return err
	}
	changedAttachments := make([]pouchdb.Attachment, 0)
	for _, pouchAtt := range *attachments {
		att, ok := model.Attachments[pouchAtt.Name]
		if ok {
			oldMd5, err := base64.StdEncoding.DecodeString(att.MD5[4:])
			if err != nil {
				return err
			}
			buf := new(bytes.Buffer)
			buf.ReadFrom(pouchAtt.Body)
			newMd5 := md5.Sum(buf.Bytes())

			if bytes.Equal(newMd5[:], oldMd5) {
				// This attachment has not changed
				continue
			}
			// The attachment has been updated, so restore the pouch attachment body
			buf.Reset()
			pouchAtt.Body = buf
		}
		changed = true
		changedAttachments = append(changedAttachments, pouchAtt)
	}

	if !changed {
		fmt.Printf("No changes to model %d / %s, skipping\n", m.Id, model.Id)
		return nil
	}

	model.Modified = m.Modified
	model.AnkiImported = time.Now()
	model.Comment = "Imported from Anki on " + time.Now().String()

	fmt.Printf("Now applying changes to model %d / %s\n", m.Id, model.Id)

	db := util.UserDb()

	rev, err := db.Put(model)
	if err != nil {
		return fmt.Errorf("Error updating model: %s", err)
	}
	for _, a := range changedAttachments {
		rev, err = db.PutAttachment(model.Id, &a, rev)
		if err != nil {
			return fmt.Errorf("Error updating model attachment: %s", err)
		}
	}
	if err != nil {
		return err
	}

	return nil
}

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
			if note.Modified.After(note.AnkiImported) {
				fmt.Printf("Note %d / %s has local changes, skipping\n", n.Id, note.Id)
				continue
			}
			fmt.Printf("Gonna update note %d / %s\n", n.Id, note.Id)
			if note.ModelId != modelUuid {
				return nil, errors.New("Changing a note's model is not supported")
			}
			if err := updateNote(&note, n, mediaMap); err != nil {
				return nil, err
			}
			continue
		}
		note = data.Note{
			Id:           noteUuid,
			Rev:          note.Rev,
			Type:         "note",
			ModelId:      modelUuid,
			Created:      note.Created,
			Modified:     n.Modified,
			AnkiImported: time.Now(),
			Tags:         n.Tags,
			FieldValues:  n.Fields,
			Comment:      "Imported from Anki on " + time.Now().String(),
		}
		rev, err := db.Put(note)
		if err != nil {
			return nil, err
		}

		attachments, err := noteAttachments(n)
		if err != nil {
			return nil, err
		}

		for _, att := range *attachments {
			filename := att.Name
			rc, err := mediaMap[filename].Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			att.Body = rc
			rev, err = db.PutAttachment(note.Id, &att, rev)
		}
	}
	return noteMap, nil
}

func noteAttachments(n *anki.Note) (*[]pouchdb.Attachment, error) {
	attachments := make([]pouchdb.Attachment, 0)

	files := make(map[string][]string)
	files["audio"] = make([]string, 0)
	files["image"] = make([]string, 0)
	for _, field := range n.Fields {
		for _, match := range soundRe.FindAllStringSubmatch(field, -1) {
			files["audio"] = append(files["audio"], match[1])
		}
		for _, match := range imageRe.FindAllStringSubmatch(field, -1) {
			files["image"] = append(files["image"], match[1])
		}
	}

	for ftype, filenames := range files {
		for _, file := range filenames {
			ext := strings.TrimPrefix(filepath.Ext(file), ".")
			contentType, ok := contentTypeMap[ext]
			if !ok {
				fmt.Printf("Unknown content type for file '%s'/%s\n", file, ext)
				contentType = ftype + "/" + ext
			}
			attachments = append(attachments, pouchdb.Attachment{
				Name: file,
				Type: contentType,
			})
		}
	}
	return &attachments, nil
}

func updateNote(note *data.Note, n *anki.Note, mediaMap map[string]*zip.File) error {
	var changed bool
	// FIXME: This is giving occasional false positives, nil vs. empty array I guess
	if !reflect.DeepEqual(n.Tags, note.Tags) {
		fmt.Printf("Tags changed. Old: %s (%d), New: %s (%d)\n", note.Tags, len(note.Tags), n.Tags, len(n.Tags))
		note.Tags = n.Tags
		changed = true
	}
	if !reflect.DeepEqual(n.Fields, note.FieldValues) {
		fmt.Printf("Fields changes. Old: %s, New: %s\n", note.FieldValues, n.Fields)
		note.FieldValues = n.Fields
		changed = true
	}

	attachments, err := noteAttachments(n)
	if err != nil {
		return err
	}

	deletedAttachments := make(map[string]bool)
	for filename, _ := range note.Attachments {
		deletedAttachments[filename] = true
	}
	changedAttachments := make([]pouchdb.Attachment, 0)
	for _, pouchAtt := range *attachments {
		delete(deletedAttachments, pouchAtt.Name)
		att, ok := note.Attachments[pouchAtt.Name]
		if ok {
			if att.Type != pouchAtt.Type {
				fmt.Printf("Content type change for %s\n", pouchAtt.Name)
			} else {
				oldMd5, err := base64.StdEncoding.DecodeString(att.MD5[4:])
				if err != nil {
					return err
				}
				rc, err := mediaMap[pouchAtt.Name].Open()
				if err != nil {
					return err
				}
				defer rc.Close()
				buf := new(bytes.Buffer)
				buf.ReadFrom(rc)
				newMd5 := md5.Sum(buf.Bytes())
				if bytes.Equal(newMd5[:], oldMd5) {
					// This attachment has not changed
					continue
				}
				fmt.Printf("MD5s differ. Old: %x, New: %x\n", oldMd5, newMd5)
			}
		}
		fmt.Printf("Attachment changed: %s\n", pouchAtt.Name)
		changed = true
		rc, err := mediaMap[pouchAtt.Name].Open()
		if err != nil {
			return err
		}
		pouchAtt.Body = rc
		changedAttachments = append(changedAttachments, pouchAtt)
	}

	for filename, _ := range deletedAttachments {
		fmt.Printf("Need to delete attachment %s\n", filename)
	}

	if !changed {
		fmt.Printf("No changes to note %d / %s, skipping\n", n.Id, note.Id)
		return nil
	}

	note.Modified = n.Modified
	note.AnkiImported = time.Now()
	note.Comment = "Imported from Anki on " + time.Now().String()

	fmt.Printf("Now applying changes to note %d / %s\n", n.Id, note.Id)

	db := util.UserDb()

	rev, err := db.Put(note)
	if err != nil {
		return fmt.Errorf("Error updating note: %s", err)
	}
	for _, a := range changedAttachments {
		rev, err = db.PutAttachment(note.Id, &a, rev)
		if err != nil {
			return fmt.Errorf("Error updating note attachment: %s", err)
		}
	}
	fmt.Printf("Updated note successfully\n")
	return nil
}

var contentTypeMap = map[string]string{
	"3gp":  "audio/3gpp",
	"ogg":  "audio/ogg",
	"mp3":  "audio/mpeg",
	"spx":  "audio/ogg",
	"wav":  "audio/x-wav",
	"flac": "audio/flac",
	"m4a":  "audio/m4a",
	"jpg":  "image/jpeg",
	"jpeg": "image/jpeg",
	"png":  "image/png",
	"gif":  "image/gif",
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
