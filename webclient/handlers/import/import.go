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
	"net/url"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/flimzy/flashback/anki"
	"github.com/flimzy/flashback/data"
	"github.com/flimzy/flashback/model/note"
	"github.com/flimzy/flashback/model/theme"
	"github.com/flimzy/flashback/util"
	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/web/file"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

var jQuery = jquery.NewJQuery

func BeforeTransition(event *jquery.Event, ui *js.Object, p url.Values) bool {
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
		return fmt.Errorf("Cannot read zip file: %s", err)
	}
	zipMap := make(map[string]*zip.File)
	for _, file := range z.File {
		zipMap[file.FileHeader.Name] = file
	}
	mediaMap, err := extractMediaMap(zipMap)
	if err != nil {
		return fmt.Errorf("Cannot extract media info: %s", err)
	}
	collection, err := extractCollection(zipMap)
	if err != nil {
		return fmt.Errorf("Cannot extract collection data: %s", err)
	}
	themeMap, err := storeModels(collection)
	if err != nil {
		return fmt.Errorf("Cannot store model: %s", err)
	}
	noteMap, err := storeNotes(collection, themeMap, mediaMap)
	if err != nil {
		return fmt.Errorf("Cannot store notes: %s", err)
	}
	deckMap, err := storeDecks(collection)
	if err != nil {
		return fmt.Errorf("Cannot store decks: %s", err)
	}
	cardMap, err := storeCards(collection, deckMap, noteMap)
	if err != nil {
		return fmt.Errorf("Cannot store cards: %s", err)
	}
	if err := storeReviews(collection, cardMap); err != nil {
		return fmt.Errorf("Cannot store reviews: %s", err)
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
{{ $g := . }}
[[- range $i, $Name := . ]]
	<div class="question" data-id="[[ $i ]]">
		{{template "[[ $Name ]] front.html" $g}}
	</div>
	<div class="answer" data-id="[[ $i ]]">
		{{template "[[ $Name ]] back.html" $g}}
	</div>
[[ end -]]
`))

type idmap map[int64]string
type thememap map[int64]*theme.Theme

var nameToIdRE = regexp.MustCompile("[[:space:]]")

func storeModels(c *anki.Collection) (thememap, error) {
	themeMap := make(thememap)
	//	db := util.UserDb()
	for _, m := range c.Models {
		if m.Type == anki.ModelTypeCloze {
			fmt.Printf("Cloze Models not yet supported\n")
			continue
		}
		fmt.Printf("Attempting to import a model\n")
		if t, err := theme.ImportAnkiModel(m); err != nil {
			return themeMap, err
		} else {
			themeMap[m.ID] = t
		}
	}
	return themeMap, nil
}

func storeDecks(c *anki.Collection) (idmap, error) {
	return nil, nil
	// 	deckMap := make(idmap)
	// 	db := util.UserDb()
	// 	for _, d := range c.Decks {
	// 		deckId := d.AnkiId()
	// 		deckMap[d.Id] = deckId
	// 		var deck data.Deck
	// 		if err := db.Get(deckId, &deck, pouchdb.Options{}); err == nil {
	// 			if deck.AnkiImported == nil || deck.Modified.After(*deck.AnkiImported) {
	// 				continue
	// 			}
	// 			if err := updateDeck(&deck, d); err != nil {
	// 				return nil, err
	// 			}
	// 			continue
	// 		}
	// 		now := time.Now()
	// 		deck = data.Deck{
	// 			Id:          deckId,
	// 			Name:        d.Name,
	// 			Description: d.Description,
	// 			Type:        "deck",
	// 			Modified:    &now,
	// 			Created:     &now,
	// 			Comment:     "Imported from Anki on " + now.String(),
	// 		}
	// 		_, err := db.Put(deck)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		if err := storeDeckConfig(c, d.ConfigId, deckId); err != nil {
	// 			return nil, err
	// 		}
	// 	}
	// 	return deckMap, nil
}

func updateDeck(deck *data.Deck, d *anki.Deck) error {
	var changed bool

	if deck.Name != d.Name {
		deck.Name = d.Name
		changed = true
	}

	if deck.Description != d.Description {
		deck.Description = d.Description
		changed = true
	}

	if !changed {
		return nil
	}

	now := time.Now()
	deck.Modified = d.Modified
	deck.AnkiImported = &now
	deck.Comment = "Imported from Anki on " + now.String()

	return updateDoc(deck, deck.Id, nil, nil)
}

func updateDoc(doc interface{}, id string, chAtt *[]pouchdb.Attachment, delAtt *map[string]bool) error {
	db, err := util.UserDb()
	if err != nil {
		return err
	}
	rev, err := db.Put(doc)
	if err != nil {
		return fmt.Errorf("Error updating doc %s: %s", id, err)
	}
	for _, a := range *chAtt {
		if rev, err = db.PutAttachment(id, &a, rev); err != nil {
			return fmt.Errorf("Error updating attachment %s for %s: %s", a.Name, id, err)
		}
	}
	for a, _ := range *delAtt {
		if rev, err = db.DeleteAttachment(id, a, rev); err != nil {
			return fmt.Errorf("Error deleting attachment %s for %s: %s", a, id, err)
		}
	}
	return nil
}

func storeDeckConfig(c *anki.Collection, dcId int64, deckId string) error {
	return nil
	// 	db := util.UserDb()
	// 	for _, dc := range c.DeckConfig {
	// 		if dc.Id == dcId {
	// 			confId := "deckconf-" + deckId
	// 			var conf data.DeckConfig
	// 			if err := db.Get(confId, &conf, pouchdb.Options{}); err == nil {
	// 				if conf.AnkiImported == nil || conf.Modified.After(*conf.AnkiImported) {
	// 					continue
	// 				}
	// 				if err := updateDeckConfig(&conf, dc); err != nil {
	// 					return err
	// 				}
	// 				continue
	// 			}
	// 			now := time.Now()
	// 			conf = data.DeckConfig{
	// 				Id:              confId,
	// 				Type:            "deckconf",
	// 				DeckId:          deckId,
	// 				Modified:        &now,
	// 				Created:         &now,
	// 				MaxDailyReviews: dc.Reviews.PerDay,
	// 				MaxDailyNew:     dc.New.PerDay,
	// 			}
	// 			if _, err := db.Put(conf); err != nil {
	// 				return err
	// 			}
	// 		}
	// 		// There's only ever one DeckConfig per Deck, so return as soon as
	// 		// we find it
	// 		return nil
	// 	}
	// 	return nil
}

func updateDeckConfig(conf *data.DeckConfig, dc *anki.DeckConfig) error {
	changed := false

	if conf.MaxDailyReviews != dc.Reviews.PerDay {
		conf.MaxDailyReviews = dc.Reviews.PerDay
		changed = true
	}

	if conf.MaxDailyNew != dc.New.PerDay {
		conf.MaxDailyNew = dc.New.PerDay
		changed = true
	}

	if !changed {
		return nil
	}

	conf.Modified = dc.Modified
	now := time.Now()
	conf.AnkiImported = &now

	return updateDoc(conf, conf.Id, nil, nil)
}

type notemap map[int64]*note.Note

var soundRe = regexp.MustCompile("\\[sound:(.*?)\\]")
var imageRe = regexp.MustCompile("<img src=\"(.*?)\" />")

func storeNotes(c *anki.Collection, themeMap thememap, mediaMap map[string]*zip.File) (notemap, error) {
	noteMap := make(notemap)
	for _, n := range c.Notes {
		t, ok := themeMap[n.ModelID]
		if !ok {
			// This isn't necessarily an error, since we skip cloze models, so just ignore for now
			continue
			// 			return nil, fmt.Errorf("Found note (id=%d) with no model", n.Id)
		}
		if note, err := note.ImportAnkiNote(t, n); err != nil {
			fmt.Printf("Error importing note: %s\n", err)
			return noteMap, err
		} else {
			noteMap[n.ID] = note
		}
	}
	return noteMap, nil
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

type cardmap map[int64]string

func storeCards(c *anki.Collection, deckMap idmap, noteMap notemap) (cardmap, error) {
	return nil, nil
	// 	cardmap := make(map[int64]string)
	// 	related := make(map[string]*[]string)
	// 	var cards []data.Card
	// 	db := util.UserDb()
	// 	for _, c := range c.Cards {
	// 		var note noteNode
	// 		var deckId string
	// 		var ok bool
	// 		if deckId, ok = deckMap[c.DeckId]; !ok {
	// 			return nil, errors.New("Found card that doesn't belong to a deck")
	// 		}
	// 		if note, ok = noteMap[c.NoteId]; !ok {
	// 			// Probably due to cloze
	// 			continue
	// 		}
	// 		cardId := c.AnkiId(note.Id)
	// 		cardmap[c.Id] = cardId
	// 		if rel, ok := related[note.Id]; ok {
	// 			*rel = append(*rel, cardId)
	// 		} else {
	// 			related[note.Id] = &([]string{cardId})
	// 		}
	// 		now := time.Now()
	// 		card := data.Card{
	// 			Id:         cardId,
	// 			Type:       "card",
	// 			NoteId:     note.Id,
	// 			DeckId:     deckId,
	// 			Modified:   &now,
	// 			Created:    &now,
	// 			Due:        c.Due,
	// 			Reviews:    c.Reps,
	// 			Lapses:     c.Lapses,
	// 			Interval:   c.Interval,
	// 			SRSFactor:  c.Factor,
	// 			TemplateId: fmt.Sprintf("%s/%d", note.Model, c.Ord),
	// 			Suspended:  c.Queue == "suspended",
	// 		}
	// 		var existing data.Card
	// 		if err := db.Get(cardId, &existing, pouchdb.Options{}); err == nil {
	// 			if err := updateCard(&existing, &card); err != nil {
	// 				return nil, err
	// 			}
	// 			continue
	// 		} else if !pouchdb.IsNotExist(err) {
	// 			fmt.Printf("Error checking for duplicate card: %s\n", err)
	// 			return nil, err
	// 		}
	// 		cards = append(cards, card)
	// 	}
	// 	for _, card := range cards {
	// 		rel := related[card.NoteId]
	// 		card.RelatedCards = make([]string, len(*rel)-1)
	// 		var i int
	// 		for _, r := range *rel {
	// 			if r != card.Id {
	// 				card.RelatedCards[i] = r
	// 				i++
	// 			}
	// 		}
	// 		if _, err := db.Put(card); err != nil {
	// 			return nil, err
	// 		}
	// 	}
	// 	return cardmap, nil
}

func updateCard(doc, card *data.Card) error {
	if doc.AnkiImported == nil || doc.Modified.After(*doc.AnkiImported) {
		return nil
	}

	changed := false

	if doc.NoteId != card.NoteId {
		return fmt.Errorf("Cannot change noteid of card %s\n", doc.Id)
	}

	if doc.TemplateId != card.TemplateId {
		return fmt.Errorf("Cannot change template of card %s\n", doc.Id)
	}

	if doc.DeckId != card.DeckId {
		doc.DeckId = card.DeckId
		changed = true
	}

	if !doc.Due.Equal(*card.Due) {
		doc.Due = card.Due
		changed = true
	}

	if doc.Reviews != card.Reviews {
		doc.Reviews = card.Reviews
		changed = true
	}

	if doc.Lapses != card.Lapses {
		doc.Lapses = card.Lapses
		changed = true
	}

	if doc.Interval != card.Interval {
		doc.Interval = card.Interval
		changed = true
	}

	if doc.SRSFactor != card.SRSFactor {
		doc.SRSFactor = card.SRSFactor
		changed = true
	}

	if doc.Suspended != card.Suspended {
		doc.Suspended = card.Suspended
		changed = true
	}

	if !reflect.DeepEqual(doc.RelatedCards, card.RelatedCards) {
		doc.RelatedCards = card.RelatedCards
		changed = true
	}

	if !changed {
		return nil
	}

	doc.Modified = card.Modified
	now := time.Now()
	doc.AnkiImported = &now

	return updateDoc(card, card.Id, nil, nil)
}

func storeReviews(c *anki.Collection, cardMap cardmap) error {
	for _, r := range c.Revlog {
		review := &data.Review{
			Id:           r.AnkiId(cardMap[r.CardId]),
			Type:         "review",
			CardId:       cardMap[r.CardId],
			Timestamp:    r.Timestamp,
			Answer:       r.Ease,
			Interval:     r.Interval,
			LastInterval: r.LastInterval,
			Factor:       r.Factor,
			ReviewType:   r.Type,
		}
		// Ignore conflict errors; it should be possible to re-import the same
		// revlog without problem
		if err := util.LogReview(review); err != nil && !pouchdb.IsConflict(err) {
			return err
		}
	}
	return nil
}
