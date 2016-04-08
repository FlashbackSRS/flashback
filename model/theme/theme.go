package theme

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"time"

	"github.com/pborman/uuid"

	"github.com/flimzy/go-pouchdb"

	"github.com/flimzy/flashback/anki"
	"github.com/flimzy/flashback/model"
	"github.com/flimzy/flashback/model/user"
)

const HTMLTemplateContentType = "text/html+flashbacktmpl"

type Attachment struct {
	ContentType string `json:"content-type"`
	Data        []byte `json:"data"`
}

type themeDoc struct {
	ID          string                       `json:"_id"`
	Rev         string                       `json:"_rev,omitempty"`
	Type        string                       `json:"type"`
	Name        *string                      `json:"name"`
	Owner       string                       `json:"owner"`
	Description *string                      `json:"description,omitempty"`
	Created     *time.Time                   `json:"created,omitempty"`
	Modified    *time.Time                   `json:"modified"`
	Imported    *time.Time                   `json:"imported,omitempty"`
	Models      []*Model                     `json:"models"`
	Attachments map[string]*model.Attachment `json:"_attachments"`
}

type Theme struct {
	doc         themeDoc
	owner       *user.User
	Name        string
	Description string
}

func New(owner *user.User) *Theme {
	t := Theme{
		owner: owner,
	}
	t.newThemeDoc()
	now := time.Now()
	t.doc.Modified = &now
	t.doc.Created = &now
	return &t
}

func (t *Theme) setID(subtype string, id uuid.UUID) {
	var buf bytes.Buffer
	buf.Write(t.owner.UUID())
	buf.Write(id)
	prefix := "theme-"
	if subtype != "" {
		prefix = prefix + subtype + "-"
	}
	t.doc.ID = prefix + hex.EncodeToString(buf.Bytes())
}

func (t *Theme) newThemeDoc() {
	t.doc = themeDoc{
		Type:        "theme",
		Name:        &t.Name,
		Description: &t.Description,
		Models:      make([]*Model, 0),
		Attachments: make(map[string]*model.Attachment),
	}
	if t.owner != nil {
		t.doc.Owner = t.owner.UUID().String()
	}
}

// ImportAnkiModel processes and stores an Anki model, updating any existing
// copy, if appropriate.
func ImportAnkiModel(m *anki.Model) error {
	u, err := user.CurrentUser()
	if err != nil {
		fmt.Printf("User not logged in\n")
		return err
	}
	now := time.Now()
	t := New(u)
	t.setID("anki", m.AnkiID())
	t.Name = m.Name
	t.doc.Created = nil
	t.doc.Modified = m.Modified
	t.doc.Imported = &now
	if m.CSS != "" {
		if err := t.AddAttachment("$main.css", "text/css", []byte(m.CSS)); err != nil {
			return err
		}
	}
	// Add the template
	tName := "$template.0.hmtl"
	thisT := &Model{Name: m.Name, Filenames: []string{tName}}
	t.doc.Models = append(t.doc.Models, thisT)
	tmpls := make([]string, len(m.Templates))
	for i, tmpl := range m.Templates {
		qName := "!" + m.Name + "." + tmpl.Name + " question.html"
		aName := "!" + m.Name + "." + tmpl.Name + " answer.html"
		if err := t.AddAttachment(qName, HTMLTemplateContentType, []byte(tmpl.QuestionFormat)); err != nil {
			return err
		}
		if err := t.AddAttachment(aName, HTMLTemplateContentType, []byte(tmpl.AnswerFormat)); err != nil {
			return err
		}
		thisT.Filenames = append(thisT.Filenames, qName, aName)
		tmpls[i] = t.Name
	}
	buf := new(bytes.Buffer)
	if err := masterTmpl.Execute(buf, tmpls); err != nil {
		return err
	}
	t.AddAttachment(tName, HTMLTemplateContentType, buf.Bytes())
	if err := t.Save(); err != nil {
		fmt.Printf("Error with first save: %s\n", err)
		if pouchdb.IsConflict(err) {
			fmt.Printf("it was a conflict error\n")
			existing, err2 := FetchTheme(t.ID())
			if err2 != nil {
				fmt.Printf("Error fetching existing doc: %s\n", err2)
				return fmt.Errorf("Fetching theme: %s\n", err2)
			}
			if err := existing.MergeImport(t); err != nil {
				fmt.Printf("merge failed: %s\n", err)
				return err
			}
			if err := existing.Save(); err != nil {
				fmt.Printf("Second save failed: %s\n", err)
				return err
			}
		} else {
			fmt.Printf("Error saving theme: %s\n", err)
			return err
		}
	}
	fmt.Printf("Save was finally successful\n")
	return nil
}

var masterTmpl = template.Must(template.New("template.html").Delims("[[", "]]").Parse(`
{{ $g := . }}
[[- range $i, $Name := . ]]
	<div class="question" data-id="[[ $i ]]">
		{{template "[[ $Name ]] question.html" $g}}
	</div>
	<div class="answer" data-id="[[ $i ]]">
		{{template "[[ $Name ]] answer.html" $g}}
	</div>
[[ end -]]
`))

func (t *Theme) AddAttachment(name, ctype string, body []byte) error {
	if name[0] != '$' && name[0] != '!' {
		return errors.New("File name must begin with $ or !")
	}
	t.doc.Attachments[name] = &model.Attachment{
		ContentType: ctype,
		Data:        body,
	}
	return nil
}

// Read-only getters

func (t *Theme) ID() string {
	return t.doc.ID
}

func (t *Theme) Rev() string {
	return t.doc.Rev
}

func (t *Theme) Created() *time.Time {
	if t.doc.Created == nil {
		return nil
	}
	ts := *t.doc.Created
	return &ts
}

func (t *Theme) Modified() *time.Time {
	if t.doc.Modified == nil {
		return nil
	}
	ts := *t.doc.Modified
	return &ts
}

func (t *Theme) Imported() *time.Time {
	if t.doc.Imported == nil {
		return nil
	}
	ts := *t.doc.Imported
	return &ts
}

func (t *Theme) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.doc)
}

func (t *Theme) UnmarshalJSON(data []byte) error {
	t.newThemeDoc()
	return json.Unmarshal(data, &t.doc)
}

func (t *Theme) stub() *model.Stub {
	return &model.Stub{
		ID:         t.doc.ID,
		Type:       "stub",
		ParentType: "theme",
	}
}

func (t *Theme) Save() error {
	u, err := user.CurrentUser()
	if err != nil {
		fmt.Printf("No current user\n")
		return err
	}
	db := pouchdb.New(t.ID())
	if rev, err := db.Put(t); err != nil {
		fmt.Printf("Doc: %v", t)
		b, e := json.Marshal(t)
		fmt.Printf("JSON err: %s\n", e)
		fmt.Printf("JSON: %s\n", b)
		return err
	} else {
		t.doc.Rev = rev
	}
	udb := pouchdb.New(u.DBName())
	s := t.stub()
	if _, err := udb.Put(s); err != nil && !pouchdb.IsConflict(err) {
		fmt.Printf("Error saving stub: %s\n", err)
		return err
	}
	return nil
}

func FetchTheme(id string) (*Theme, error) {
	db := pouchdb.New(id)
	t := &Theme{}
	err := db.Get(id, t, pouchdb.Options{})
	return t, err
}

func (t *Theme) MergeImport(n *Theme) error {
	if t.Imported() == nil {
		return errors.New("Conflict. Cannot MergeImport to a non-imported theme")
	}
	if t.Modified().After(*n.Imported()) {
		return errors.New("The theme has been modified since last import. Merge not possible.")
	}
	t.Name = n.Name
	t.Description = n.Description
	t.doc.Modified = n.doc.Modified
	t.doc.Imported = n.doc.Imported
	t.doc.Models = n.doc.Models
	t.doc.Attachments = n.doc.Attachments
	return nil
}
