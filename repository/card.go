package repo

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/davecgh/go-spew/spew"
	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/log"
	"github.com/pkg/errors"
	"golang.org/x/net/html"

	"github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/cardmodel"
	"github.com/FlashbackSRS/flashback/util"
)

// Card provides a convenient interface to fb.Card and dependencies
type Card struct {
	*fb.Card
	db   *DB
	note *Note
}

type jsCard struct {
	ID string `json:"id"`
}

// MarshalJSON marshals a Card for the benefit of javascript context in HTML
// templates.
func (c *Card) MarshalJSON() ([]byte, error) {
	card := &jsCard{
		ID: c.DocID(),
	}
	return json.Marshal(card)
}

// Note returns the card's associated Note
func (c *Card) Note() (*Note, error) {
	if err := c.fetchNote(); err != nil {
		return nil, errors.Wrap(err, "Error fetching note for Note()")
	}
	return c.note, nil
}

func (c *Card) fetchNote() error {
	if c.note != nil {
		// Nothing to do
		return nil
	}
	log.Debugf("Fetching note %s", c.NoteID())
	db, err := NewDB(c.BundleID())
	if err != nil {
		return errors.Wrap(err, "fetchNote() can't connect to bundle DB")
	}
	n := &fb.Note{}
	if err := db.Get(c.NoteID(), n, pouchdb.Options{Attachments: true}); err != nil {
		return errors.Wrapf(err, "fetchNote() can't fetch %s", c.NoteID())
	}
	c.note = &Note{
		Note: n,
		db:   db,
	}
	return nil
}

// GetCard fetches the requested card
func (u *User) GetCard(id string) (*Card, error) {
	db, err := u.DB()
	if err != nil {
		return nil, errors.Wrap(err, "Unable to connect to User DB")
	}

	card := &fb.Card{}
	if err := db.Get(id, card, pouchdb.Options{}); err != nil {
		return nil, errors.Wrap(err, "Unable to fetch requested card")
	}
	return &Card{
		Card: card,
		db:   db,
	}, nil
}

type cardPriority struct {
	Card     *fb.Card
	Priority float32
}

type prioritizedCards []cardPriority

func (p prioritizedCards) Len() int { return len(p) }
func (p prioritizedCards) Less(i, j int) bool {
	return p[i].Priority > p[j].Priority || p[i].Card.Created.Before(p[j].Card.Created)
}
func (p prioritizedCards) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// CardPrio returns a number 0 or greater, as a priority to be used in
// determining card study order.
func CardPrio(due time.Time, interval time.Duration, now time.Time) float32 {
	return float32(math.Pow(1+float64(now.Sub(due))/float64(interval), 3))
}

// GetCards fetches up to max cards from the db, in priority order.
func GetCards(db *DB, now time.Time, max int) ([]*fb.Card, error) {
	doc := make(map[string][]*fb.Card)
	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"type":    "card",
			"due":     map[string]interface{}{"$gte": nil},
			"created": map[string]interface{}{"$gte": nil},
		},
		//		"fields": []string{"_id", "due", "interval", "model", "created"},
		"sort":  []string{"due", "created"},
		"limit": 100,
	}
	// spew.Dump(query)
	if err := db.Find(query, &doc); err != nil {
		return nil, errors.Wrap(err, "card list")
	}
	// spew.Dump(doc)
	fmt.Printf("10\n")
	pri := make([]cardPriority, len(doc["docs"]))
	fmt.Printf("20\n")
	for i, card := range doc["docs"] {
		fmt.Printf("30: %d\n", i)
		pri[i].Card = card
		if card.Due != nil {
			pri[i].Priority = CardPrio(*card.Due, *card.Interval, now)
		}
		spew.Dump(pri[i])
	}
	fmt.Printf("40\n")
	sort.Sort(prioritizedCards(pri))
	docs := make([]*fb.Card, len(pri))
	for i, card := range pri {
		docs[i] = card.Card
	}
	return docs, nil
}

// GetNextCard gets the next card to study
func (u *User) GetNextCard() (*Card, error) {
	db, err := u.DB()
	if err != nil {
		return nil, errors.Wrap(err, "GetNextCard(): Error connecting to User DB")
	}
	cl, err := GetCards(db, time.Now(), 10)
	if err != nil {
		return nil, err
	}
	spew.Dump(cl)

	doc := make(map[string][]*fb.Card)
	query := map[string]interface{}{
		"selector": map[string]string{"type": "card"},
		"fields":   []string{"_id", "due", "interval", "model"},
		"sort":     "due",
		"limit":    100,
	}
	if err := db.Find(query, &doc); err != nil {
		return nil, errors.Wrap(err, "GetNextCard(): Error fetching card")
	}
	spew.Dump(doc)
	return nil, nil
	if len(doc["docs"]) == 0 {
		return nil, errors.New("No cards available")
	}
	return &Card{
		Card: doc["docs"][0],
		db:   db,
	}, nil
}

type cardContext struct {
	IframeID string
	Card     *Card
	Note     *Note
	// Model    *Model
	// Deck     *Deck
	BaseURI string
	Fields  map[string]template.HTML
}

const (
	// Question is a card's first face
	Question = iota
	// Answer is a card's second face
	Answer
)

var faces = map[int]string{
	Question: "question",
	Answer:   "answer",
}

// ModelHandler returns the cardmodel.Model for this card
// FIXME: Rename this method to just Model() (??)
func (c *Card) ModelHandler() (cardmodel.Model, error) {
	m, err := c.Model()
	if err != nil {
		return nil, errors.Wrap(err, "retrieve model")
	}
	return cardmodel.GetHandler(m.Type)
}

// Model returns the model for the card
func (c *Card) Model() (*Model, error) {
	note, err := c.Note()
	if err != nil {
		return nil, errors.Wrap(err, "retrieve Note")
	}
	model, err := note.Model()
	if err != nil {
		return nil, errors.Wrap(err, "retrieve Model")
	}
	return model, nil
}

// Body returns the requested card face
func (c *Card) Body(face int) (body string, iframeID string, err error) {
	note, err := c.Note()
	if err != nil {
		return "", "", errors.Wrap(err, "Unable to retrieve Note")
	}
	model, err := c.Model()
	if err != nil {
		return "", "", errors.Wrap(err, "Unable to retrieve Model")
	}
	tmpl, err := model.GenerateTemplate()
	if err != nil {
		return "", "", errors.Wrap(err, "Error generating template")
	}
	ctx := cardContext{
		IframeID: RandString(8),
		Card:     c,
		Note:     note,
		// Model:    model,
		BaseURI: util.BaseURI(),
		Fields:  make(map[string]template.HTML),
	}

	for i, f := range model.Fields {
		switch note.FieldValues[i].Type() {
		case fb.AnkiField, fb.TextField:
			text, e := note.FieldValues[i].Text()
			if e != nil {
				return "", "", errors.Wrap(e, "Unable to fetch text for field value")
			}
			ctx.Fields[f.Name] = template.HTML(text)
		}
	}

	htmlDoc := new(bytes.Buffer)
	if e := tmpl.Execute(htmlDoc, ctx); e != nil {
		return "", "", errors.Wrap(e, "Unable to execute template")
	}
	log.Debugf("original size = %d\n", htmlDoc.Len())
	newBody, err := prepareBody(face, c.TemplateID(), model.Type, htmlDoc)
	if err != nil {
		return "", "", errors.Wrap(err, "prepare body")
	}

	nbString := string(newBody)
	log.Debugf("new body size = %d\n", len(nbString))
	return nbString, ctx.IframeID, nil
}

func prepareBody(face int, templateID uint32, modelType string, r io.Reader) ([]byte, error) {
	cardFace, ok := faces[face]
	if !ok {
		return nil, errors.Errorf("Unrecognized card face %d", face)
	}
	handler, err := cardmodel.GetHandler(modelType)
	if err != nil {
		return nil, errors.Wrap(err, "model handler")
	}
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, errors.Wrap(err, "goquery parse")
	}
	body := doc.Find("body")
	if body == nil {
		return nil, errors.New("no body in template output")
	}
	sel := fmt.Sprintf("div.%s[data-id='%d']", cardFace, templateID)
	container := body.Find(sel)
	if container.Length() == 0 {
		return nil, errors.Errorf("No div matching '%s' found in template output", sel)
	}

	containerHTML, err := container.Html()
	if err != nil {
		return nil, errors.Wrap(err, "error extracting div html")
	}

	body.Empty()
	body.AppendHtml(containerHTML)

	doc.Find("head").AppendHtml(fmt.Sprintf(`<script type="text/javascript">%s</script>`, string(handler.IframeScript())))

	newBody, err := goquery.OuterHtml(doc.Selection)
	if err != nil {
		return nil, errors.Wrap(err, "outer html failed")
	}
	return []byte(newBody), nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// Random number function borrowed from http://stackoverflow.com/a/31832326/13860
var src = rand.NewSource(time.Now().UnixNano())

// RandString returns a random string of n bytes, converted to hex
func RandString(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return hex.EncodeToString(b)
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

func findContainer(n *html.Node, targetID, targetClass string) *html.Node {
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
		if class == targetClass && id == targetID {
			return n
		}
	}
	return findContainer(n.NextSibling, targetID, targetClass)
}

// GetAttachment fetches an attachment from the note, failling back to the model
func (c *Card) GetAttachment(filename string) (*Attachment, error) {
	n, err := c.Note()
	if err != nil {
		return nil, errors.Wrap(err, "Error fetching Note for GetAttachment()")
	}
	if file, ok := n.Attachments.GetFile(filename); ok {
		return &Attachment{file}, nil
	}

	m, err := n.Model()
	if err != nil {
		return nil, errors.Wrap(err, "Error fetching Model for GetAttachments()")
	}
	if file, ok := m.Files.GetFile(filename); ok {
		return &Attachment{file}, nil
	}
	return nil, errors.Errorf("File '%s' not found", filename)
}

// Response represents a response button
type Response struct {
	Name    string
	Display string
	Icon    string
}

var showAnswer = &Response{
	Name:    "show_answer_button",
	Display: "Show Answer",
	Icon:    "carat-r",
}

var wrongAnswer = &Response{
	Name:    "wrong_answer_button",
	Display: "Again",
	Icon:    "delete",
}

var hardAnswer = &Response{
	Name:    "hard_answer_button",
	Display: "Hard",
	Icon:    "clock",
}

var goodAnswer = &Response{
	Name:    "good_answer_button",
	Display: "Good",
	Icon:    "carat-r",
}

var easyAnswer = &Response{
	Name:    "easy_answer_button",
	Display: "Easy",
	Icon:    "heart",
}

// Responses returns the list of available responses for a card's face
func (c *Card) Responses(face int) ([]*Response, error) {
	var responses []*Response
	switch face {
	case Question:
		responses = []*Response{showAnswer}
	case Answer:
		responses = []*Response{wrongAnswer, hardAnswer, goodAnswer, easyAnswer}
	default:
		return nil, errors.Errorf("Unknown card face %d", face)
	}
	return responses, nil
}
