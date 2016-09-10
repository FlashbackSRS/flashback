package repo

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"html/template"
	"math/rand"
	"time"

	"golang.org/x/net/html"

	"github.com/flimzy/go-pouchdb"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/util"
)

// Card provides a convenient interface to fb.Card and dependencies
type Card struct {
	*fb.Card
	db   *DB
	note *Note
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
		db:   c.db,
	}
	return nil
}

/*
func (c *Card) modelID() (string, int, error) {
	parts := strings.Split(c.Identity(), ".")
	if len(parts) != 3 {
		return "", 0, errors.New("Invalid card ID: " + c.Identity())
	}
	modelID, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", 0, errors.Wrap(err, "Unable to parse ModelID")
	}
	return "theme-" + parts[1], modelID, nil
}
*/

// GetCard isn't currently used (???) FIXME
func GetCard() (*Card, error) {
	c := &Card{}
	if err := c.fetchArbitraryCard(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Card) fetchArbitraryCard() error {
	u, err := CurrentUser()
	if err != nil {
		return err
	}
	if c.db == nil {
		db, err := u.DB()
		if err != nil {
			return err
		}
		c.db = db
	}
	doc := make(map[string][]*fb.Card)
	query := map[string]interface{}{
		"selector": map[string]string{"type": "card"},
		"limit":    1,
	}
	if err := c.db.Find(query, &doc); err != nil {
		return err
	}
	if len(doc["docs"]) == 0 {
		return errors.New("No cards available")
	}
	c.Card = doc["docs"][0]
	return nil
}

type cardContext struct {
	IframeID string
	Card     *Card
	Note     *Note
	Model    *Model
	Deck     *Deck
	BaseURI  string
	Fields   map[string]template.HTML
}

// Body returns the card's body and iframe ID
func (c *Card) Body() (string, string, error) {
	note, err := c.Note()
	if err != nil {
		return "", "", errors.Wrap(err, "Unable to retrieve Note")
	}
	model, err := note.Model()
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
		Model:    model,
		BaseURI:  util.BaseURI(),
		Fields:   make(map[string]template.HTML),
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
	doc, err := html.Parse(htmlDoc)
	if err != nil {
		return "", "", errors.Wrap(err, "Unable to parse generated HTML")
	}
	body := findBody(doc)
	if body == nil {
		return "", "", errors.New("No <body> in the template output")
	}
	fmt.Printf("%s\n", htmlDoc)
	/*
		container := findContainer(body.FirstChild, c.ModelID(), "question")
		if container == nil {
			return "", "", errors.New("No matching div found in template output")
		}
		// Delete unused divs
		for c := body.FirstChild; c != nil; c = body.FirstChild {
			body.RemoveChild(c)
		}
		inner := container.FirstChild
		inner.Parent = body
		body.FirstChild = inner

		newBody := new(bytes.Buffer)
		if err := html.Render(newBody, doc); err != nil {
			return "", "", errors.Wrap(err, "Error rendering new HTML")
		}
		nbString := newBody.String()
		fmt.Printf("original size = %d\n", len(htmlDoc.String()))
		fmt.Printf("new body size = %d\n", len(nbString))
		return nbString, ctx.IframeID, nil
	*/
	return "", "", nil
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
