package anki

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*
 * https://gist.github.com/sartak/3921255
 * https://github.com/ankidroid/Anki-Android/wiki/Database-Structure
 */

// CREATE TABLE col (
//     id              integer primary key, -- Arbitrary number, since there's only one row
//     crt             integer not null,
//     mod             integer not null,
//     scm             integer not null,
//     ver             integer not null,
//     dty             integer not null, -- No longer used: https://github.com/dae/anki/blob/master/anki/collection.py#L90
//     usn             integer not null, -- update sequence number : used to figure out diffs when syncing, not useful for us
//     ls              integer not null,
//     conf            text not null,
//     models          text not null,
//     decks           text not null,
//     dconf           text not null,
//     tags            text not null -- a cache of tags used in the collection, not useful for us
// );

type Collection struct {
	Created        *time.Time
	Modified       *time.Time
	SchemaModified *time.Time
	Ver            int
	LastSync       *time.Time
	Config         *Config
	Models         []*Model
	Decks          []*Deck
	DeckConfig     []*DeckConfig
	Cards          []*Card
	Notes          []*Note
	Revlog         []*Review
}

// defaultConf = {
//     # review options
//     'activeDecks': [1],
//     'curDeck': 1,
//     'newSpread': NEW_CARDS_DISTRIBUTE,
//     'collapseTime': 1200,
//     'timeLim': 0,
//     'estTimes': True,
//     'dueCounts': True,
//     # other config
//     'curModel': None,
//     'nextPos': 1,
//     'sortType': "noteFld",
//     'sortBackwards': False,
//     'addToCur': True, # add new to currently selected deck?
// }
// # Also found
// newBury -- not used?

func jsMilliseconds(ms int64) *time.Time {
	t := time.Unix(int64(ms/1000), int64(ms%1000))
	return &t
}

func jsSeconds(s int64) *time.Time {
	t := time.Unix(s, 0)
	return &t
}

func SqliteToCollection(row map[string]interface{}) (*Collection, error) {
	c := &Collection{
		Created:        jsSeconds(int64(row["crt"].(float64))),
		Modified:       jsMilliseconds(int64(row["mod"].(float64))),
		SchemaModified: jsMilliseconds(int64(row["scm"].(float64))),
		Ver:            int(row["ver"].(float64)),
		LastSync:       jsMilliseconds(int64(row["ls"].(float64))),
	}
	if err := c.parseConfig(row["conf"].(string)); err != nil {
		return nil, fmt.Errorf("[conf] %s", err)
	}
	if err := c.parseModels(row["models"].(string)); err != nil {
		return nil, fmt.Errorf("[models] %s", err)
	}
	if err := c.parseDecks(row["decks"].(string)); err != nil {
		return nil, fmt.Errorf("[decks] %s", err)
	}
	if err := c.parseDeckConfig(row["dconf"].(string)); err != nil {
		return nil, fmt.Errorf("[dconf] %s", err)
	}
	return c, nil
}

type Config struct {
	ActiveDecks   []int `json:"activeDecks"`
	CurrentDeck   int64 `json:"curDeck"`
	NewSpread     int   `json:"newSpread"`
	CollapseTime  int   `json:"collapseTime"`
	timeLimit     int   `json:"timeLim"`
	TimeLimit     time.Duration
	EstTimes      bool   `json:"estTimes"`
	DueCounts     bool   `json:"dueCounts"`
	SortType      string `json:"sortType"`
	SortBackwards bool   `json:"sortBackwards"`
	// 	NextPos       uint   `json:"nextPos"`
	// 	CurrentModel  int64   `json:"curModel"`
}

func (c *Collection) parseConfig(jsonString string) error {
	var conf Config
	if err := json.Unmarshal([]byte(jsonString), &conf); err != nil {
		return err
	}
	conf.TimeLimit = time.Duration(int64(conf.timeLimit)) * time.Millisecond
	c.Config = &conf
	return nil
}

type ModelType uint

const (
	ModelTypeStandard ModelType = iota
	ModelTypeCloze
)

// AKA "Note Type"
type Model struct {
	Id        int64       `json:"-"`
	Name      string      `json:"name"`
	Tags      []string    `json:"tags"`
	DeckId    int64       `json:"did"`
	Fields    []*Field    `json:"flds"`
	SortField int         `json:"sortf"`
	Templates []*Template `json:"tmpls"`
	Type      ModelType   `json:"type"`
	LatexPre  string      `json:"latexPre"`
	LatexPost string      `json:"latexPost"`
	CSS       string      `json:"css"`
	Mod       int64       `json:"mod"`
	Modified  *time.Time
	// Req string `json:"req"` -- Required fields? Possibly auto-generated after examining templates?
}

func int64ToBytes(id int64) []byte {
	buf := make([]byte, 6)
	binary.PutVarint(buf, id)
	return buf
}

func b64(buf []byte) string {
	return strings.Replace(base64.StdEncoding.EncodeToString(buf), "/", "_", -1)
}

func (m *Model) AnkiID() []byte {
	var buf bytes.Buffer
	buf.Write(int64ToBytes(m.Id))
	buf.Write([]byte{byte(m.Type)})
	return buf.Bytes()
}

type Field struct {
	Name     string `json:"name"`
	Sticky   bool   `json:"sticky"`
	RTL      bool   `json:"rtl"`
	Ord      int    `json:"ord"`
	Font     string `json:"font"`
	FontSize int    `json:"size"`
}

// AKA "Card Type"
type Template struct {
	Name           string `json:"name"`
	QuestionFormat string `json:"qfmt"`
	AnswerFormat   string `json:"afmt"`
}

func (c *Collection) parseModels(jsonString string) error {
	var models map[string]Model
	if err := json.Unmarshal([]byte(jsonString), &models); err != nil {
		return err
	}
	for i, m := range models {
		if id, err := strconv.ParseInt(i, 10, 64); err != nil {
			return err
		} else {
			m.Id = id
		}
		m.Modified = jsSeconds(m.Mod)
		for _, t := range m.Templates {
			a, err := convertTemplate(t.AnswerFormat)
			if err != nil {
				return err
			}
			q, err := convertTemplate(t.QuestionFormat)
			if err != nil {
				return err
			}
			t.AnswerFormat = a
			t.QuestionFormat = q
		}
		c.Models = append(c.Models, &m)
	}
	return nil
}

/*
Anki templates may contian the following types of tags:
{{Name}} -- Simple variable substitution
{{type:Name}} -- Typing dialog
{{cloze:Name}} -- Cloze replacement
{{hint:Name}} -- Hint field
{{#Name}} -- "If defined"
{{/Name}} -- End "if defined"
{{text:Name}} -- Variable subsititution, removing HTML markup
*/

var tagRe *regexp.Regexp = regexp.MustCompile("{{.*?}}")

func convertTemplate(ankiTmpl string) (string, error) {
	var converted bytes.Buffer
	var i = 0

	tags := tagRe.FindAllStringIndex(ankiTmpl, -1)

	for _, tag := range tags {
		content := strings.Trim(ankiTmpl[tag[0]:tag[1]], "{ }")
		if tag[0] > i {
			converted.WriteString(ankiTmpl[i:tag[0]])
		}
		i = tag[1]

		switch {
		case strings.HasPrefix(content, "type:"):
			converted.WriteString("{{/* " + content + " */}}")
		case strings.HasPrefix(content, "cloze:"):
			converted.WriteString("{{/* " + content + " */}}")
		case strings.HasPrefix(content, "hint:"):
			converted.WriteString("{{/* " + content + " */}}")
		case strings.HasPrefix(content, "text:"): // Strips HTML tags
			converted.WriteString("{{/* " + content + " */}}")
		case strings.HasPrefix(content, "#"): // Defined
			converted.WriteString("{{/* " + content + " */}}")
		case strings.HasPrefix(content, "/"): // End
			converted.WriteString("{{/* " + content + " */}}")
		case strings.HasPrefix(content, "^"): // Not defined
			converted.WriteString("{{/* " + content + " */}}")
		default:
			converted.WriteString("{{ ." + content + " }}")
		}
	}
	if i < len(ankiTmpl) {
		converted.WriteString(ankiTmpl[i:])
	}

	return string(converted.Bytes()), nil
}

type Deck struct {
	Id               int64      `json:"-"`
	Name             string     `json:"name"`
	Mid              string     `json:"mid"`
	ModelId          int64      `json:"-"`
	Description      string     `json:"descr"`
	ExtendedRev      uint8      `json:"extendedRev"`
	Collapsed        bool       `json:"collapsed"`
	BrowserCollapsed bool       `json:"browserCollapsed"`
	NewToday         []int      `json:"newToday"`
	timeToday        []int      `json:"timeToday"`
	Dyn              int        `json:"dyn"`
	ExtendedNew      int        `json:"extendedNew"`
	ConfigId         int64      `json:"conf"`
	ReviewToday      []int      `json:"revToday"`
	LearnToday       []int      `json:"lrnToday"`
	Mod              int64      `json:"mod"`
	Modified         *time.Time `json:"-"`
}

func (d Deck) AnkiId() string {
	return "deck-anki-" + b64(int64ToBytes(d.Id))
}

func (c *Collection) parseDecks(jsonString string) error {
	var decks map[string]Deck
	if err := json.Unmarshal([]byte(jsonString), &decks); err != nil {
		return err
	}
	for i, d := range decks {
		if id, err := strconv.ParseInt(i, 10, 64); err != nil {
			return err
		} else {
			d.Id = id
		}
		if d.Mid != "" {
			if mid, err := strconv.ParseInt(d.Mid, 10, 64); err != nil {
				return err
			} else {
				d.ModelId = mid
			}
		}
		d.Modified = jsSeconds(d.Mod)
		c.Decks = append(c.Decks, &d)
	}
	return nil
}

type DeckConfig struct {
	Id       int64      `json:"-"`
	Name     string     `json:"name"`
	ReplayQ  bool       `json:"replayq"`
	Timer    int        `json:"timer"`
	MaxTaken int        `json:"maxTaken"`
	Mod      int64      `json:"mod"`
	Modified *time.Time `json:"-"`
	Autoplay bool       `json:"autoplay"`
	Lapses   struct {
		LeechFails  int   `json:"leechFails"`
		MinInt      int   `json:"minInt"`
		Delays      []int `json:"delays"`
		LeechAction int   `json:"leechAction"`
		Mult        int   `json:"mult"`
	} `json:"lapse"`
	Reviews struct {
		PerDay      int           `json:"perDay"`
		Fuzz        float32       `json:"fuzz"`
		IntervalFct int           `json:"ivlFct"`
		MaxInterval time.Duration `json:"maxIvl"`
		Ease4       float32       `json:"ease4"`
		Bury        bool          `json:"bury"`
		MinSpace    int           `json:"minSpace"`
	} `json:"rev"`
	New struct {
		PerDay        int             `json:"perDay"`
		Delays        []int           `json:"delays"`
		Separate      bool            `json:"separate"`
		Intervals     []time.Duration `json:"ints"`
		InitialFactor float32         `json:"initialFactor"`
		Bury          bool            `json:"bury"`
		Order         int             `json:"order"`
	} `json:"new"`
}

func (c *Collection) parseDeckConfig(jsonString string) error {
	var dconf map[string]DeckConfig
	if err := json.Unmarshal([]byte(jsonString), &dconf); err != nil {
		return err
	}
	for i, dc := range dconf {
		if id, err := strconv.ParseInt(i, 10, 64); err != nil {
			return err
		} else {
			dc.Id = id
		}
		dc.Modified = jsSeconds(dc.Mod)
		c.DeckConfig = append(c.DeckConfig, &dc)
	}
	return nil
}

// CREATE TABLE cards (
//     id              integer primary key,   /* 0 */
//     nid             integer not null,      /* 1 */
//     did             integer not null,      /* 2 */
//     ord             integer not null,      /* 3 */
//     mod             integer not null,      /* 4 */
//     usn             integer not null,      /* 5 */ -- update sequence number : used to figure out diffs when syncing, not useful for us
//     type            integer not null,      /* 6 */
//     queue           integer not null,      /* 7 */
//     due             integer not null,      /* 8 */
//     ivl             integer not null,      /* 9 */
//     factor          integer not null,      /* 10 */
//     reps            integer not null,      /* 11 */
//     lapses          integer not null,      /* 12 */
//     left            integer not null,      /* 13 */ -- reps left till graduation
//     odue            integer not null,      /* 14 */ -- original due: only used when the card is currently in filtered deck, not useful for us
//     odid            integer not null,      /* 15 */ -- original did: only used when the card is currently in filtered deck, not useful for us
//     flags           integer not null,      /* 16 */ -- no longer used
//     data            text not null          /* 17 */ -- unused
// );
// CREATE INDEX ix_cards_usn on cards (usn);
// CREATE INDEX ix_cards_nid on cards (nid);
// CREATE INDEX ix_cards_sched on cards (did, queue, due);

type Card struct {
	Id       int64
	NoteId   int64
	DeckId   int64
	Ord      int
	Modified *time.Time
	Type     string
	Queue    string
	Due      *time.Time
	Interval time.Duration
	Factor   float32
	Reps     int
	Lapses   int
	Left     int
}

var CardTypes = map[int]string{
	0: "new",
	1: "learning",
	2: "due",
}

var QueueTypes = map[int]string{
	0: "suspended",
	1: "userburied",
	2: "schedburied",
}

func (c Card) AnkiId(note string) string {
	note = strings.TrimPrefix(note, "note-anki-")
	return fmt.Sprintf("card-anki-%s-%s", b64(int64ToBytes(c.Id)), note)
}

func (c *Collection) AddCardFromSqlite(row map[string]interface{}) {
	card := &Card{
		Id:       int64(row["id"].(float64)),
		NoteId:   int64(row["nid"].(float64)),
		DeckId:   int64(row["did"].(float64)),
		Ord:      int(row["ord"].(float64)),
		Modified: jsSeconds(int64(row["ord"].(float64))),
		Type:     CardTypes[int(row["type"].(float64))],
		Queue:    QueueTypes[int(row["queue"].(float64))],
		Interval: time.Duration(int64(row["ivl"].(float64))) * time.Hour * 24,
		Factor:   float32(row["factor"].(float64)) / 1000,
		Reps:     int(row["reps"].(float64)),
		Lapses:   int(row["lapses"].(float64)),
		Left:     int(row["left"].(float64)),
	}
	switch card.Type {
	case "new":
		// Due dates make no sense for new cards
	case "due":
		t := c.Created.AddDate(0, 0, int(row["due"].(float64)))
		card.Due = &t
	case "learning":
		card.Due = jsSeconds(int64(row["due"].(float64)))
	}
	c.Cards = append(c.Cards, card)
}

// CREATE TABLE notes (
//     id              integer primary key,   /* 0 */
//     guid            text not null,         /* 1 */
//     mid             integer not null,      /* 2 */
//     mod             integer not null,      /* 3 */
//     usn             integer not null,      /* 4 */ -- update sequence number : used to figure out diffs when syncing, not useful for us
//     tags            text not null,         /* 5 */
//     flds            text not null,         /* 6 */ -- the values of the fields in this note. separated by 0x1f (31) character.
//     sfld            integer not null,      /* 7 */
//     csum            integer not null,      /* 8 */
//     flags           integer not null,      /* 9 */ -- unused
//     data            text not null          /* 10 */ -- unused
// );
// CREATE INDEX ix_notes_usn on notes (usn);
// CREATE INDEX ix_notes_csum on notes (csum);

type Note struct {
	Id        int64
	Guid      string
	ModelId   int64
	Modified  *time.Time
	Tags      []string
	Fields    []string
	SortField string
	Csum      int64
}

func (n Note) AnkiId() string {
	return fmt.Sprintf("note-anki-%s-%s", b64(int64ToBytes(n.Id)), b64([]byte(n.Guid)))
}

func (c *Collection) AddNoteFromSqlite(row map[string]interface{}) {
	c.Notes = append(c.Notes, &Note{
		Id:        int64(row["id"].(float64)),
		Guid:      row["guid"].(string),
		ModelId:   int64(row["mid"].(float64)),
		Modified:  jsSeconds(int64(row["mod"].(float64))),
		Tags:      strings.Fields(row["tags"].(string)),
		Fields:    strings.Split(row["flds"].(string), "\x1f"),
		SortField: row["sfld"].(string),
		Csum:      int64(row["csum"].(float64)),
	})
}

// CREATE TABLE revlog (
//     id              integer primary key,
//     cid             integer not null,
//     usn             integer not null, -- update sequence number : used to figure out diffs when syncing, not useful for us
//     ease            integer not null,
//     ivl             integer not null, -- negative = seconds, positive = days
//     lastIvl         integer not null, -- Same as ivl
//     factor          integer not null,
//     time            integer not null,
//     type            integer not null  -- 0=lrn, 1=rev, 2=relrn, 3=cram
// );
// CREATE INDEX ix_revlog_usn on revlog (usn);
// CREATE INDEX ix_revlog_cid on revlog (cid);

type Review struct {
	Timestamp    *time.Time
	CardId       int64
	Ease         string
	Interval     time.Duration
	LastInterval time.Duration
	Factor       float32
	ReviewTime   time.Duration
	Type         string
}

var EaseTypes = map[int]string{
	0: "wrong",
	1: "hard",
	2: "ok",
	3: "easy",
}

var ReviewTypes = map[int]string{
	0: "learn",
	1: "review",
	2: "relearn",
	3: "cram",
}

func (r Review) AnkiId(cid string) string {
	cardId := strings.TrimPrefix(cid, "card-anki-")
	return fmt.Sprintf("review-%s-%s", cardId, b64(int64ToBytes(r.Timestamp.Unix())))
}

func (c *Collection) AddReviewFromSqlite(row map[string]interface{}) {
	c.Revlog = append(c.Revlog, &Review{
		Timestamp:    jsMilliseconds(int64(row["id"].(float64))),
		CardId:       int64(row["cid"].(float64)),
		Ease:         EaseTypes[int(row["ease"].(float64))],
		Interval:     intervalToDuration(int(row["ivl"].(float64))),
		LastInterval: intervalToDuration(int(row["lastIvl"].(float64))),
		Factor:       float32(row["factor"].(float64)) / 1000,
		ReviewTime:   time.Duration(row["time"].(float64)) * time.Millisecond,
		Type:         ReviewTypes[int(row["type"].(float64))],
	})
}

func intervalToDuration(interval int) time.Duration {
	if interval < 0 {
		return -time.Duration(int64(interval)) * time.Second
	}
	return time.Duration(int64(interval)) * time.Hour * 24
}

func (c *Collection) DeleteDeck(id int64) {
}
