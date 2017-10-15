package model

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/flimzy/kivik"
	"github.com/flimzy/log"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback"
	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/controllers/done"
	"github.com/FlashbackSRS/flashback/controllers/mustsync"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
)

// Card wraps an *fb.Card and its dependencies.
type Card struct {
	*fb.Card
	note   *fbNote
	model  *fbModel
	appURL string
	repo   *Repo
}

var _ flashback.CardView = &Card{}

type jsCard struct {
	ID      string      `json:"id"`
	ModelID int         `json:"model"`
	Context interface{} `json:"context,omitempty"`
}

// MarshalJSON marshals a Card for the benefit of javascript context in HTML
// templates.
func (c *Card) MarshalJSON() ([]byte, error) {
	card := &jsCard{
		ID:      c.ID,
		ModelID: c.ThemeModelID(),
		Context: c.Context,
	}
	return json.Marshal(card)
}

// Buttons returns the list of buttons to be made visible for the card face.
func (c *Card) Buttons(face int) (studyview.ButtonMap, error) {
	mc, err := GetModelController(c.model.Type)
	if err != nil {
		return nil, err
	}
	return mc.Buttons(face)
}

type cardData struct {
	Card *Card
	Face int
	Note *fbNote
	// Model    *Model
	// Deck     *Deck
	BaseURI string
	Fields  map[string]template.HTML
}

// Body produces the HTML body of the card to be displayed.
func (c *Card) Body(ctx context.Context, face int) (body string, err error) {
	defer profile("Body")()
	cardFace, ok := faces[face]
	if !ok {
		return "", errors.Errorf("unrecognized card face %d", face)
	}
	if c.note == nil {
		return "", errors.New("card hasn't been fetched")
	}

	funcMap, err := c.model.FuncMap(c, face)
	if err != nil {
		return "", errors.Wrap(err, "failed to get FuncMap")
	}

	tmpl, err := c.model.Template(ctx)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate template")
	}

	data := cardData{
		Card: c,
		Face: face,
		Note: c.note,
		// Model:    model,
		BaseURI: c.appURL,
		Fields:  make(map[string]template.HTML, len(c.model.Fields)),
	}

	for i, field := range c.model.Fields {
		switch c.note.FieldValues[i].Type() {
		case fb.AnkiField, fb.TextField:
			data.Fields[field.Name] = template.HTML(c.note.FieldValues[i].Text)
		}
	}

	htmlDoc := new(bytes.Buffer)
	if e := tmpl.Funcs(funcMap).Execute(htmlDoc, data); e != nil {
		return "", errors.Wrap(e, "template execution")
	}
	// return htmlDoc.String(), nil

	iframeScript, _ := c.model.IframeScript()
	newBody, err := prepareBody(cardFace, c.TemplateID(), string(iframeScript), htmlDoc)
	if err != nil {
		return "", errors.Wrap(err, "prepare body")
	}

	nbString := string(newBody)
	log.Debugf("new body size = %d\n", len(nbString))
	return nbString, nil
}

// Action handles a card action produced by the user.
func (c *Card) Action(ctx context.Context, face *int, startTime time.Time, query interface{}) (done bool, err error) {
	mc, err := GetModelController(c.model.Type)
	if err != nil {
		return false, err
	}
	done, err = mc.Action(c, face, startTime, query)
	if err != nil {
		return false, err
	}
	db, err := c.repo.userDB(ctx)
	if err != nil {
		return false, err
	}
	return done, saveDoc(ctx, db, c.Card)
}

var now = time.Now

// The priority for new cards.
const newPriority = 0.5

// batch sizes are the number of cards we fetch at once, using simple schedule
// prioritization. This number should be large enough that the intelligent
// scheduling has room to function, but small enough that performance isn't
// a big problem due to fetching and prioritizing many cards we don't actually
// use.
const (
	newBatchSize = 5
	oldBatchSize = 45

	// limitPadding is added to the query limit, to reduce the total number of
	// queries which must be performed.
	limitPadding = 20
)

func getCardsFromView(ctx context.Context, db querier, view, deck string, limit int) ([]*cardSchedule, error) {
	defer profile("getCardsFromView: " + view)()
	if limit <= 0 {
		return nil, errors.New("invalid limit")
	}
	cards := make([]*cardSchedule, 0, limit)
	offset := 0
	for i := 0; len(cards) < limit && i < 100; i++ {
		result, readRows, err := queryView(ctx, db, view, deck, limit, offset)
		if err != nil {
			return nil, err
		}
		if len(result) > 0 {
			cards = append(cards, result...)
		}
		if len(cards) == limit || readRows < limit {
			break
		}
		offset = offset + readRows
	}
	return cards, nil
}

const (
	mainDDoc = "index"
	mainView = "cards"
)

func queryView(ctx context.Context, db querier, state, deck string, limit, offset int) (cards []*cardSchedule, readRows int, err error) {
	defer profile("queryView: " + state)()
	log.Debugf("Trying to fetch %d (%d) %s cards\n", limit, offset, state)
	query := map[string]interface{}{
		"limit":    limit + limitPadding,
		"skip":     offset,
		"reduce":   false,
		"startkey": []interface{}{state, deck},
		"endkey":   []interface{}{state, deck, map[string]interface{}{}},
	}
	rows, err := db.Query(ctx, mainDDoc, mainView, query)
	if err != nil {
		return nil, 0, errors.Wrap(err, "query failed")
	}
	defer func() { _ = rows.Close() }()
	cards = make([]*cardSchedule, 0, limit)
	var count int
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		count++
		card := &cardSchedule{}
		if e := rows.ScanValue(card); e != nil {
			return nil, count, errors.Wrap(e, "ScanValue")
		}
		if card.BuriedUntil.After(fb.Due(now())) {
			continue
		}
		var key []string
		if e := rows.ScanKey(&key); e != nil {
			return nil, count, errors.Wrap(e, "ScanKey")
		}
		due, err := dueFromKey(key)
		if err != nil {
			return nil, count, errors.Wrap(err, "due time")
		}
		card.Due = due
		card.ID = rows.ID()
		cards = append(cards, card)
		if len(cards) == limit {
			log.Debugf("Got %d cards, early exiting", len(cards))
			return cards, count, nil
		}
	}
	return cards, count, nil
}

func dueFromKey(key []string) (fb.Due, error) {
	var raw string
	switch len(key) {
	case 2:
		// legacy index
		raw = key[0]
	case 3:
		// new index
		raw = key[1]
	case 4:
		raw = key[2]
	default:
		return fb.Due{}, fmt.Errorf("Key has %d element(s), expected 2, 3 or 4", len(key))
	}
	if raw != "" {
		return fb.ParseDue(raw)
	}
	return fb.Due{}, nil
}

// cardPriority returns a number 0 or greater, as a priority to be used in
// determining card study order.
func cardPriority(due fb.Due, interval fb.Interval, now time.Time) float64 {
	if due.IsZero() || interval == 0 {
		return newPriority
	}
	// Remove the timezone
	_, offset := now.Zone()
	utc := now.UTC().Add(time.Duration(offset) * time.Second)

	return float64(math.Pow(1+float64(utc.Sub(time.Time(due)))/float64(time.Duration(interval)), 3))
}

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func selectWeightedCard(cards []*cardSchedule) string {
	switch len(cards) {
	case 0:
		return ""
	case 1:
		return cards[0].ID
	}
	var weights float64
	priorities := make([]float64, len(cards))
	for i, card := range cards {
		priority := cardPriority(card.Due, card.Interval, now())
		priorities[i] = priority
		weights += priority
	}
	r := rnd.Float64() * weights
	log.Debugf("Selected r=%f of %f\n", r, weights)
	for i, priority := range priorities {
		r -= priority
		if r < 0 {
			log.Debugf("Selected card %d: %s (prio: %f, %0.2f%% chance)\n", i, cards[i].ID, priority, priority/weights*100)
			return cards[i].ID
		}
	}
	// should never happen
	return ""
}

// GetCardToStudy returns a CardView to display to the user to study, and buries
// related cards.
func (r *Repo) GetCardToStudy(ctx context.Context, deck string) (flashback.CardView, error) {
	if _, _, err := r.lastSyncTime(ctx); err != nil {
		if kivik.StatusCode(err) == kivik.StatusNotFound {
			return mustsync.GetCard(), nil
		}
		return nil, err
	}
	card, err := r.getCardToStudy(ctx, deck)
	if err != nil {
		return nil, err
	}
	if err == nil && card == nil {
		return done.GetCard(), nil
	}
	go func() {
		// Bury related cards
		if err := r.BuryRelatedCards(ctx, card.Card); err != nil {
			log.Printf("Failed to bury cards: %s\n", err)
		}
	}()

	return card, nil
}

// getCardToStudy returns a card to display to the user to study.
func (r *Repo) getCardToStudy(ctx context.Context, deck string) (*Card, error) {
	udb, err := r.userDB(ctx)
	if err != nil {
		return nil, err
	}
	card, err := getCardToStudy(ctx, udb, deck)
	if err != nil || card == nil {
		return nil, err
	}
	c := &Card{
		Card:   card,
		appURL: r.appURL,
		repo:   r,
	}
	return c, c.fetch(ctx, r.local)
}

func (c *Card) fetch(ctx context.Context, client kivikClient) error {
	if c.note != nil {
		return nil
	}
	db, err := client.DB(ctx, c.BundleID())
	if err != nil {
		return err
	}
	note := &fb.Note{}
	theme := &fb.Theme{}
	var noteErr, themeErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		noteErr = getDoc(ctx, db, c.NoteID(), &note)
		wg.Done()
	}()
	go func() {
		themeErr = getDoc(ctx, db, c.ThemeID(), &theme)
		wg.Done()
	}()
	wg.Wait()
	if err := firstErr(noteErr, themeErr); err != nil {
		return err
	}
	if len(theme.Models) == 0 {
		// This means corrupted/broken data
		return errors.New("card's theme has no model")
	}
	c.note = &fbNote{Note: note}
	model := theme.Models[c.ThemeModelID()]
	c.model = &fbModel{Model: model, db: db}
	return c.note.SetModel(model)
}

type cardSchedule struct {
	ID          string      `json:"_id"`
	Interval    fb.Interval `json:"interval"`
	Due         fb.Due      `json:"due"`
	BuriedUntil fb.Due      `json:"buriedUntil"`
}

func getCardToStudy(ctx context.Context, db queryGetter, deck string) (*fb.Card, error) {
	defer profile("getCardToStudy")()
	var newCards, oldCards []*cardSchedule
	var newErr, oldErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		newCards, newErr = getCardsFromView(ctx, db, "new", deck, newBatchSize)
		newErr = errors.Wrap(newErr, "new")
		wg.Done()
	}()
	go func() {
		oldCards, oldErr = getCardsFromView(ctx, db, "old", deck, oldBatchSize)
		oldErr = errors.Wrap(oldErr, "old")
		wg.Done()
	}()
	wg.Wait()
	if err := firstErr(newErr, oldErr); err != nil {
		return nil, err
	}
	cardID := selectWeightedCard(append(newCards, oldCards...))
	if cardID == "" {
		return nil, nil
	}
	row, err := db.Get(ctx, cardID)
	if err != nil {
		return nil, err
	}
	card := &fb.Card{}
	err = row.ScanDoc(&card)
	return card, err
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

func prepareBody(cardFace string, templateID uint32, iframeScript string, r io.Reader) ([]byte, error) {
	defer profile("prepareBody")()
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, errors.Wrap(err, "goquery parse")
	}
	body := doc.Find("body")
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
	body.AppendHtml(fmt.Sprintf(`<form id="mainform">%s</form>`, containerHTML))
	body.AddClass("card", fmt.Sprintf("card%d", templateID+1))

	doc.Find("head").AppendHtml(fmt.Sprintf(`<script type="text/javascript">%s</script>`, iframeScript))

	newBody, err := goquery.OuterHtml(doc.Selection)
	if err != nil {
		return nil, errors.Wrap(err, "outer html failed")
	}
	return []byte(newBody), nil
}

func relatedKeyRange(cardID string) (startKey, endKey string) {
	startKey = strings.TrimRight(cardID, "0123456789")
	return startKey, startKey + string(rune(0x10FFFF))
}

// Fields returns a list of field names for the card.
func (c *Card) Fields() []string {
	fields := make([]string, len(c.model.Fields))
	for i, field := range c.model.Fields {
		fields[i] = field.Name
	}
	return fields
}

// FieldValue returns the card's value for the requested field.
func (c *Card) FieldValue(fieldName string) *fb.FieldValue {
	for i, field := range c.model.Fields {
		if field.Name == fieldName {
			return c.note.FieldValues[i]
		}
	}
	return nil
}
