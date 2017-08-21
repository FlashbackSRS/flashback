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
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/flimzy/log"
	"github.com/pkg/errors"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/webclient/views/studyview"
)

// Card represents a generic card-like object.
type Card interface {
	DocID() string
	Buttons(face int) (studyview.ButtonMap, error)
	Body(face int) (body string, err error)
	Action(face *int, startTime time.Time, query interface{}) (done bool, err error)
	BuryRelated() error
}

// fbCard is a wrapper around *fb.Card, which provides convenience functions
// like .Note(), .Model(), and .Body()
type fbCard struct {
	*fb.Card
	note   *fbNote
	model  *fbModel
	appURL string
}

var _ Card = &fbCard{}

type jsCard struct {
	ID      string      `json:"id"`
	ModelID int         `json:"model"`
	Context interface{} `json:"context,omitempty"`
}

// MarshalJSON marshals a Card for the benefit of javascript context in HTML
// templates.
func (c *fbCard) MarshalJSON() ([]byte, error) {
	card := &jsCard{
		ID:      c.ID,
		ModelID: c.ThemeModelID(),
		Context: c.Context,
	}
	return json.Marshal(card)
}

func (c *fbCard) Buttons(face int) (studyview.ButtonMap, error) {
	return studyview.ButtonMap{}, nil
}

type cardData struct {
	Card *fbCard
	Face int
	Note *fbNote
	// Model    *Model
	// Deck     *Deck
	BaseURI string
	Fields  map[string]template.HTML
}

func (c *fbCard) Body(face int) (body string, err error) {
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

	tmpl, err := c.model.Template()
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

func (c *fbCard) Action(face *int, startTime time.Time, query interface{}) (done bool, err error) {
	return false, nil
}

func (c *fbCard) BuryRelated() error {
	return nil
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
	newBatchSize = 10
	oldBatchSize = 90
)

// limitPadding is a number added to the limit parameter passed to the
// getCardsFromView function. This is added, because there's no automated way
// to eliminate buried cards from the view, so they must be filtered in the
// client, but this could lead to queries with no results, so we pad the number
// of results to help reduce this chance.
const limitPadding = 100

func getCardsFromView(ctx context.Context, db querier, view string, limit, offset int) ([]*fb.Card, error) {
	if limit <= 0 {
		return nil, errors.New("invalid limit")
	}
	rows, err := db.Query(context.TODO(), "index", view, map[string]interface{}{
		"limit":        limit + limitPadding,
		"offset":       offset,
		"include_docs": true,
	})
	if err != nil {
		return nil, errors.Wrap(err, "query failed")
	}
	defer func() { _ = rows.Close() }()
	cards := make([]*fb.Card, 0, limit)
	var count int
	for rows.Next() {
		count++
		card := &fb.Card{}
		if err := rows.ScanDoc(card); err != nil {
			return nil, err
		}
		if card.BuriedUntil.After(fb.Due(now())) {
			continue
		}
		if card.Interval != 0 {
			// Skip cards we already saw today, with an interval >= 1d; they would make no progress.
			if card.Interval.Days() >= 1 && !time.Time(fb.On(now())).After(card.LastReview) {
				continue
			}
			// Skip sub-day intervals that aren't due yet. We only allow forward-fuzzing for intervals > 1day
			if card.Interval.Days() == 0 && card.Due.After(fb.Due(now())) {
				continue
			}
		}
		cards = append(cards, card)
		if len(cards) == limit {
			return cards, nil
		}
	}
	if rows.TotalRows() > int64(limit+offset) {
		more, err := getCardsFromView(ctx, db, view, limit-len(cards), offset+count)
		return append(cards, more...), err
	}
	return cards, nil
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

func selectWeightedCard(cards []*fb.Card) *fb.Card {
	switch len(cards) {
	case 0:
		return nil
	case 1:
		return cards[0]
	}
	var weights float64
	priorities := make([]float64, len(cards))
	for i, card := range cards {
		priority := cardPriority(card.Due, card.Interval, now())
		priorities[i] = priority
		weights += priority
	}
	r := rnd.Float64() * weights
	for i, priority := range priorities {
		r -= priority
		if r < 0 {
			return cards[i]
		}
	}
	// should never happen
	return nil
}

// GetCardToStudy returns a card to display to the user to study.
func (r *Repo) GetCardToStudy(ctx context.Context) (Card, error) {
	udb, err := r.userDB(ctx)
	if err != nil {
		return nil, err
	}
	card, err := getCardToStudy(ctx, udb)
	if err != nil {
		return nil, err
	}
	c := &fbCard{
		Card:   card,
		appURL: r.appURL,
	}
	return c, c.fetch(ctx, r.local)
}

func (c *fbCard) fetch(ctx context.Context, client kivikClient) error {
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
	c.model = &fbModel{Model: model}
	return c.note.SetModel(model)
}

func getCardToStudy(ctx context.Context, db querier) (*fb.Card, error) {
	defer profile("getCardToStudy")()
	var newCards, oldCards []*fb.Card
	var newErr, oldErr error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		newCards, newErr = getCardsFromView(ctx, db, "newCards", newBatchSize, 0)
		wg.Done()
	}()
	go func() {
		oldCards, oldErr = getCardsFromView(ctx, db, "oldCards", oldBatchSize, 0)
		wg.Done()
	}()
	wg.Wait()
	if err := firstErr(newErr, oldErr); err != nil {
		return nil, err
	}
	return selectWeightedCard(append(newCards, oldCards...)), nil
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
