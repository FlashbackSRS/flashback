package repo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math"
	"math/rand"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/davecgh/go-spew/spew"
	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/pkg/errors"
	"golang.org/x/net/html"

	"github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/repository/done"
	"github.com/FlashbackSRS/flashback/util"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
)

const newPriority = 0.5

func init() {
	rand.Seed(int64(time.Now().UnixNano()))
}

// Card represents a generic card-like object.
type Card interface {
	DocID() string
	Buttons(face int) (studyview.ButtonMap, error)
	Body(face int) (body string, err error)
	Action(face *int, startTime time.Time, query *js.Object) (done bool, err error)
}

// PouchCard provides a convenient interface to fb.Card and dependencies
type PouchCard struct {
	*fb.Card
	db   *DB
	note *Note
}

type jsCard struct {
	ID      string      `json:"id"`
	ModelID int         `json:"model"`
	Context interface{} `json:"context,omitempty"`
}

// MarshalJSON marshals a Card for the benefit of javascript context in HTML
// templates.
func (c *PouchCard) MarshalJSON() ([]byte, error) {
	card := &jsCard{
		ID:      c.DocID(),
		ModelID: int(c.ModelID()),
		Context: c.Context,
	}
	return json.Marshal(card)
}

// Save saves the card's current state to the database.
func (c *PouchCard) Save() error {
	log.Debugf("Attempting to save card %s\n", c.Identity())
	return c.db.Save(c.Card)
}

// Note returns the card's associated Note
func (c *PouchCard) Note() (*Note, error) {
	if err := c.fetchNote(); err != nil {
		return nil, errors.Wrap(err, "Error fetching note for Note()")
	}
	return c.note, nil
}

func (c *PouchCard) fetchNote() error {
	if c.note != nil {
		// Nothing to do
		return nil
	}
	log.Debugf("Fetching note %s", c.NoteID())
	db, err := c.db.User.NewDB(c.BundleID())
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
func (u *User) GetCard(id string) (*PouchCard, error) {
	db, err := u.DB()
	if err != nil {
		return nil, errors.Wrap(err, "connect to user DB")
	}
	return db.GetCard(id)
}

// GetCard fetches the requested card
func (db *DB) GetCard(id string) (*PouchCard, error) {
	card := &fb.Card{}
	if err := db.Get(id, card, pouchdb.Options{}); err != nil {
		return nil, errors.Wrap(err, "fetch card")
	}
	return &PouchCard{
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
func CardPrio(due *fb.Due, interval *fb.Interval, now time.Time) float32 {
	if due == nil || interval == nil {
		return newPriority
	}
	// Remove the timezone
	_, offset := now.Zone()
	utc := now.UTC().Add(time.Duration(offset) * time.Second)

	return float32(math.Pow(1+float64(utc.Sub(time.Time(*due)))/float64(time.Duration(*interval)), 3))
}

// A CardListItem contains the minimal information necessary to determine
// card ordering.
type CardListItem struct {
	ID          string       `json:"_id"`
	Due         *fb.Due      `json:"due"`
	Created     time.Time    `json:"created"`
	Interval    *fb.Interval `json:"interval"`
	ReviewCount int          `json:"reviewCount"`
	priority    float32
}

// GetCardList returns a list, limit elements long, of cards sorted by due
// date and created time.
func GetCardList(db *DB, limit int) ([]*CardListItem, error) {
	var newCards []*CardListItem
	var oldCards []*CardListItem
	newLimit := int(float64(limit) * 0.1)
	if newLimit < 1 {
		newLimit = 1
	}
	log.Debugf("newLimit = %d\n", newLimit)
	newQuery := map[string]interface{}{
		"selector": map[string]interface{}{
			"type":      "card",
			"created":   map[string]interface{}{"$gte": nil}, // Only so we can sort on this field
			"due":       map[string]interface{}{"$eq": nil},
			"suspended": map[string]interface{}{"$ne": true},
		},
		// Sort by due, just so we use the proper index; due is always the same
		// for these results.
		"sort":   []string{"due", "created"},
		"fields": []string{"_id", "interval", "reviewCount"},
		"limit":  newLimit,
	}
	if err := db.Find(newQuery, &newCards); err != nil {
		return nil, errors.Wrap(err, "new card list")
	}

	oldLimit := limit - len(newCards)
	log.Debugf("oldLimit = %d\n", oldLimit)
	oldQuery := map[string]interface{}{
		"selector": map[string]interface{}{
			"type":      "card",
			"due":       map[string]interface{}{"$gt": nil},
			"created":   map[string]interface{}{"$gte": nil}, // Just so we use the proper index
			"suspended": map[string]interface{}{"$ne": true},
			"$not": map[string]interface{}{
				"$or": []interface{}{
					map[string]interface{}{
						// Don't review cards we already saw today, with an
						// interval >= 1d; they would make no progress.
						"interval":   map[string]interface{}{"$gte": 1},
						"lastReview": map[string]interface{}{"$gte": fb.Today().String()},
					},
					map[string]interface{}{
						// Ignore sub-day intervals that aren't due yet. We only allow
						// forward-fuzzing for intervals > 1day
						"interval": map[string]interface{}{"$lt": 0},
						"due":      map[string]interface{}{"$gt": fb.Now().String()},
					},
					map[string]interface{}{
						"suspended": true,
					},
				},
			},
		},
		"sort":   []string{"due", "created"},
		"fields": []string{"due", "_id", "interval", "reviewCount"},
		"limit":  oldLimit,
	}
	if err := db.Find(oldQuery, &oldCards); err != nil {
		return nil, errors.Wrap(err, "card list")
	}
	cards := append(oldCards, newCards...)
	if len(cards) > limit {
		cards = cards[0:limit]
	}
	log.Debugf("%d new cards, %d old cards (%d cards select in all)", len(newCards), len(oldCards), len(cards))
	return cards, nil
}

// UnmarshalJSON wraps fb.Card's Unmarshaler
func (c *PouchCard) UnmarshalJSON(data []byte) error {
	fbCard := &fb.Card{}
	err := json.Unmarshal(data, fbCard)
	c.Card = fbCard
	return err
}

// cardBatchSize is the number of cards we fetch at once, using simple schedule
// prioritization. This number should be large enough that the intelligent
// scheduling has room to function, but small enough that performance isn't
// a big problem due to fetching and prioritizing many cards we don't actually
// use.
const cardBatchSize = 100

// GetNextCard gets the next card to study
func (u *User) GetNextCard() (Card, error) {
	db, err := u.DB()
	if err != nil {
		return nil, errors.Wrap(err, "GetNextCard(): Error connecting to User DB")
	}

	cards, err := GetCardList(db, cardBatchSize)
	if err != nil {
		return nil, errors.Wrap(err, "get card list")
	}
	for _, card := range cards {
		card.priority = CardPrio(card.Due, card.Interval, time.Now())
	}
	if len(cards) == 0 {
		return done.GetCard(), nil
	}

	var weights float32
	for _, c := range cards {
		weights += c.priority
	}

	r := rand.Float32() * weights
	log.Debugf("Random key / total: %f / %f (%d)\n", r, weights, len(cards))
	for i, c := range cards {
		r -= c.priority
		if r < 0 {
			log.Debugf("Selected card %d: %s\n", i, c.ID)
			return u.GetCard(c.ID)
		}
	}
	return nil, errors.New("failed to fetch card")
}

type cardContext struct {
	Card *PouchCard
	Face int
	Note *Note
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

// Buttons returns the button states for the given card/face.
func (c *PouchCard) Buttons(face int) (studyview.ButtonMap, error) {
	cont, err := c.getModelController()
	if err != nil {
		return nil, err
	}
	return cont.Buttons(face)
}

// Action handles the action on the card, such as a button press.
func (c *PouchCard) Action(face *int, startTime time.Time, query *js.Object) (done bool, err error) {
	cont, err := c.getModelController()
	if err != nil {
		return false, err
	}
	return cont.Action(c, face, startTime, query)
}

// Model returns the model for the card
func (c *PouchCard) Model() (*Model, error) {
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
func (c *PouchCard) Body(face int) (body string, err error) {
	note, err := c.Note()
	if err != nil {
		return "", errors.Wrap(err, "Unable to retrieve Note")
	}
	model, err := c.Model()
	if err != nil {
		return "", errors.Wrap(err, "Unable to retrieve Model")
	}
	tmpl, err := model.GenerateTemplate()
	if err != nil {
		return "", errors.Wrap(err, "Error generating template")
	}
	ctx := cardContext{
		Card: c,
		Face: face,
		Note: note,
		// Model:    model,
		BaseURI: util.BaseURI(),
		Fields:  make(map[string]template.HTML),
	}

	for i, f := range model.Fields {
		switch note.FieldValues[i].Type() {
		case fb.AnkiField, fb.TextField:
			text, e := note.FieldValues[i].Text()
			if e != nil {
				return "", errors.Wrap(e, "Unable to fetch text for field value")
			}
			ctx.Fields[f.Name] = template.HTML(text)
		}
	}

	spew.Dump(ctx)

	funcs, err := model.FuncMap(face)
	if err != nil {
		return "", errors.Wrap(err, "failed to get FuncMap")
	}

	htmlDoc := new(bytes.Buffer)
	if err = tmpl.Funcs(funcs).Execute(htmlDoc, ctx); err != nil {
		return "", errors.Wrap(err, "Unable to execute template")
	}
	cont, err := c.getModelController()
	if err != nil {
		return "", errors.Wrap(err, "get model controller")
	}
	log.Debugf("original size = %d\n", htmlDoc.Len())
	newBody, err := prepareBody(face, c.TemplateID(), cont, htmlDoc)
	if err != nil {
		return "", errors.Wrap(err, "prepare body")
	}

	nbString := string(newBody)
	log.Debugf("new body size = %d\n", len(nbString))
	return nbString, nil
}

func prepareBody(face int, templateID uint32, cont ModelController, r io.Reader) ([]byte, error) {
	cardFace, ok := faces[face]
	if !ok {
		return nil, errors.Errorf("Unrecognized card face %d", face)
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
		html, _ := goquery.OuterHtml(doc.Selection)
		fmt.Printf("----------\n%s\n----------------\n", html)
		return nil, errors.Errorf("No div matching '%s' found in template output", sel)
	}

	containerHTML, err := container.Html()
	if err != nil {
		return nil, errors.Wrap(err, "error extracting div html")
	}

	body.Empty()
	body.AppendHtml(fmt.Sprintf(`<form id="mainform">%s</form>`, containerHTML))
	body.AddClass("card", fmt.Sprintf("card%d", templateID+1))

	doc.Find("head").AppendHtml(fmt.Sprintf(`<script type="text/javascript">%s</script>`, string(cont.IframeScript())))

	newBody, err := goquery.OuterHtml(doc.Selection)
	if err != nil {
		return nil, errors.Wrap(err, "outer html failed")
	}
	return []byte(newBody), nil
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
func (c *PouchCard) GetAttachment(filename string) (*Attachment, error) {
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
