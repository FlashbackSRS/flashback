package anki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"regexp"
	"strings"
	"time"
	// 	"honnef.co/go/js/console"
)

type CardType uint8

const (
	CardTypeNew CardType = iota
	CardTypeLearning
	CardTypeDue
)

type QueueType int8

const (
	QueueTypeSuspended QueueType = iota * -1
	QueueTypeUserBuried
	QueueTypeSchedBuried
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
	Created        time.Time
	Modified       time.Time
	SchemaModified time.Time
	Ver            int
	LastSync       time.Time
	Config         Config
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

func jsTime(ms uint64) time.Time {
	return time.Unix(int64(ms/1000), int64(ms%1000))
}

func SqliteToCollection(row map[string]interface{}) (*Collection, error) {
	c := &Collection{
		Created:        time.Unix(int64(row["crt"].(float64)), 0),
		Modified:       jsTime(uint64(row["mod"].(float64))),
		SchemaModified: jsTime(uint64(row["scm"].(float64))),
		Ver:            int(row["ver"].(float64)),
		LastSync:       jsTime(uint64(row["ls"].(float64))),
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
	ActiveDecks   []uint `json:"activeDecks"`
	CurrentDeck   uint   `json:"curDeck"`
	NewSpread     uint   `json:"newSpread"`
	CollapseTime  uint   `json:"collapseTime"`
	TimeLimit     uint   `json:"timeLim"`
	EstTimes      bool   `json:"estTimes"`
	DueCounts     bool   `json:"dueCounts"`
	SortType      string `json:"sortType"`
	SortBackwards bool   `json:"sortBackwards"`
	// 	NextPos       uint   `json:"nextPos"`
	// 	CurrentModel  uint   `json:"curModel"`
}

func (c *Collection) parseConfig(jsonString string) error {
	return json.Unmarshal([]byte(jsonString), &(c.Config))
}

type Model struct {
	Id        uint64     `json:"-"`
	Name      string     `json:"name"`
	Tags      []string   `json:"tags"`
	DeckId    uint64     `json:"did"`
	Fields    []Field    `json:"flds"`
	SortField uint8      `json:"sortf"`
	Templates []*Template `json:"tmpls"`
	Type      uint8      `json:"type"`
	LatexPre  string     `json:"latexPre"`
	LatexPost string     `json:"latexPost"`
	CSS       string     `json:"css"`
	Mod       int64      `json:"mod"`
	Modified  time.Time
	// Req string `json:"req"` -- Required fields? Possibly auto-generated after examining templates?
}

type Field struct {
	Name   string `json:"name"`
	Sticky bool   `json:"sticky"`
	RTL    bool   `json:"rtl"`
	Ord    int    `json:"ord"`
	Font   string `json:"font"`
	Size   int    `json:"size"`
}

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
		if id, err := strconv.ParseUint(i, 10, 64); err != nil {
			return err
		} else {
			m.Id = id
		}
		m.Modified = time.Unix(m.Mod, 0)
		for _,t := range m.Templates {
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

func convertTemplate(ankiTmpl string) (string,error) {
	var converted bytes.Buffer
	var i = 0
	
	tags := tagRe.FindAllStringIndex( ankiTmpl, -1 )
	
	for _, tag := range tags {
		content := strings.Trim(ankiTmpl[tag[0]:tag[1]], "{ }")
		fmt.Printf("Found tag: {{%s}}\n", content)
		if tag[0] > i {
			converted.WriteString( ankiTmpl[i:tag[0]] )
		}
		i = tag[1]
		
		switch {
			case strings.HasPrefix(content,"type:"):
				converted.WriteString("{{/* " + content + " */}}")
			case strings.HasPrefix(content,"cloze:"):
				converted.WriteString("{{/* " + content + " */}}")
			case strings.HasPrefix(content,"hint:"):
				converted.WriteString("{{/* " + content + " */}}")
			case strings.HasPrefix(content,"text:"):
				converted.WriteString("{{/* " + content + " */}}")
			case strings.HasPrefix(content,"#"):
				converted.WriteString("{{/* " + content + " */}}")
			case strings.HasPrefix(content,"/"):
				converted.WriteString("{{/* " + content + " */}}")
			default:
				converted.WriteString("{{ ." + content + " }}")
		}
	}
	if i < len(ankiTmpl) {
		converted.WriteString( ankiTmpl[i:] )
	}
	
	return string(converted.Bytes()), nil
}

type Deck struct {
	Id               uint64    `json:"-"`
	Name             string    `json:"name"`
	Mid              string    `json:"mid"`
	ModelId          uint64    `json:"-"`
	Description      string    `json:"descr"`
	ExtendedRev      uint8     `json:"extendedRev"`
	Collapsed        bool      `json:"collapsed"`
	BrowserCollapsed bool      `json:"browserCollapsed"`
	NewToday         []uint    `json:"newToday"`
	timeToday        []uint    `json:"timeToday"`
	Dyn              uint8     `json:"dyn"`
	ExtendedNew      uint8     `json:"extendedNew"`
	ConfigId         uint64    `json:"conf"`
	ReviewToday      []uint    `json:"revToday"`
	LearnToday       []uint    `json:"lrnToday"`
	Mod              int64     `json:"mod"`
	Modified         time.Time `json:"-"`
}

func (c *Collection) parseDecks(jsonString string) error {
	var decks map[string]Deck
	if err := json.Unmarshal([]byte(jsonString), &decks); err != nil {
		return err
	}
	for i, d := range decks {
		if id, err := strconv.ParseUint(i, 10, 64); err != nil {
			return err
		} else {
			d.Id = id
		}
		if d.Mid != "" {
			if mid, err := strconv.ParseUint(d.Mid, 10, 64); err != nil {
				return err
			} else {
				d.ModelId = mid
			}
		}
		d.Modified = time.Unix(d.Mod, 0)
		c.Decks = append(c.Decks, &d)
	}
	return nil
}

type DeckConfig struct {
	Id       uint64    `json:"-"`
	Name     string    `json:"name"`
	ReplayQ  bool      `json:"replayq"`
	Timer    uint8     `json:"timer"`
	MaxTaken uint8     `json:"maxTaken"`
	Mod      int64     `json:"mod"`
	Modified time.Time `json:"-"`
	Autoplay bool      `json:"autoplay"`
	Lapse    struct {
		LeechFails  uint   `json:"leechFails"`
		MinInt      uint   `json:"minInt"`
		Delays      []uint `json:"delays"`
		LeechAction uint8  `json:"leechAction"`
		Mult        uint   `json:"mult"`
	} `json:"lapse"`
	Rev struct {
		PerDay      uint    `json:"perDay"`
		Fuzz        float32 `json:"fuzz"`
		IntervalFct uint    `json:"ivlFct"`
		MaxInterval uint    `json:"maxIvl"`
		Ease4       float32 `json:"ease4"`
		Bury        bool    `json:"bury"`
		MinSpace    uint    `json:"minSpace"`
	} `json:"rev"`
	New struct {
		PerDay        uint   `json:"perDay"`
		Delays        []uint `json:"delays"`
		Separate      bool   `json:"separate"`
		Intervals     []uint `json:"ints"`
		InitialFactor uint   `json:"initialFactor"`
		Bury          bool   `json:"bury"`
		Order         uint   `json:"order"`
	} `json:"new"`
}

func (c *Collection) parseDeckConfig(jsonString string) error {
	var dconf map[string]DeckConfig
	if err := json.Unmarshal([]byte(jsonString), &dconf); err != nil {
		return err
	}
	for i, dc := range dconf {
		if id, err := strconv.ParseUint(i, 10, 64); err != nil {
			return err
		} else {
			dc.Id = id
		}
		dc.Modified = time.Unix(dc.Mod, 0)
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
	Id       uint64
	NoteId   uint64
	DeckId   uint64
	Ord      int
	Modified time.Time
	Type     CardType
	Queue    QueueType
	Due      int
	Interval int
	Factor   int
	Reps     uint
	Lapses   uint
	Left     int
}

func (c *Collection) AddCardFromSqlite(row map[string]interface{}) {
	c.Cards = append(c.Cards, &Card{
		Id:       uint64(row["id"].(float64)),
		NoteId:   uint64(row["nid"].(float64)),
		DeckId:   uint64(row["did"].(float64)),
		Ord:      int(row["ord"].(float64)),
		Modified: time.Unix(int64(row["ord"].(float64)), 0),
		Type:     CardType(row["type"].(float64)),
		Queue:    QueueType(row["queue"].(float64)),
		Due:      int(row["due"].(float64)),
		Interval: int(row["ivl"].(float64)),
		Factor:   int(row["factor"].(float64)),
		Reps:     uint(row["reps"].(float64)),
		Lapses:   uint(row["lapses"].(float64)),
		Left:     int(row["left"].(float64)),
	})
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
	Id        uint64
	Guid      string
	ModelId   uint64
	Modified  time.Time
	Tags      []string
	Fields    []string
	SortField string
	Csum      uint64
}

func (c *Collection) AddNoteFromSqlite(row map[string]interface{}) {
	c.Notes = append(c.Notes, &Note{
		Id:        uint64(row["id"].(float64)),
		Guid:      row["guid"].(string),
		ModelId:   uint64(row["mid"].(float64)),
		Modified:  time.Unix(int64(row["mod"].(float64)), 0),
		Tags:      strings.Fields(row["tags"].(string)),
		Fields:    strings.Split(row["flds"].(string), "\x1f"),
		SortField: row["sfld"].(string),
		Csum:      uint64(row["csum"].(float64)),
	})
}

// CREATE TABLE revlog (
//     id              integer primary key,
//     cid             integer not null,
//     usn             integer not null, -- update sequence number : used to figure out diffs when syncing, not useful for us
//     ease            integer not null,
//     ivl             integer not null,
//     lastIvl         integer not null,
//     factor          integer not null,
//     time            integer not null,
//     type            integer not null
// );
// CREATE INDEX ix_revlog_usn on revlog (usn);
// CREATE INDEX ix_revlog_cid on revlog (cid);

type Ease uint8

const (
	EaseWrong Ease = iota
	EaseHard
	EaseOK
	EaseEasy
)

type Review struct {
	Id           uint
	CardId       uint64
	Ease         Ease
	Interval     int
	LastInterval int
	Factor       int
	Time         time.Duration
	Type         int
}

func (c *Collection) AddReviewFromSqlite(row map[string]interface{}) {
	c.Revlog = append(c.Revlog, &Review{
		Id:           uint(row["id"].(float64)),
		CardId:       uint64(row["cid"].(float64)),
		Ease:         Ease(row["ease"].(float64)),
		Interval:     int(row["ivl"].(float64)),
		LastInterval: int(row["lastIvl"].(float64)),
		Factor:       int(row["factor"].(float64)),
		Time:         time.Duration(row["time"].(float64)) * time.Millisecond,
		Type:         int(row["type"].(float64)),
	})
}

func (c *Collection) DeleteDeck(id uint64) {
}
