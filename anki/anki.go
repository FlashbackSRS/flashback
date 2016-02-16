package anki

import (
	"encoding/json"
	"strconv"
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
	Models         []Model
	Cards          []Card
	Notes          []Note
	Revlog         []Review
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

type Config struct {
	ActiveDecks  []uint `json:"activeDecks"`
	CurrentDeck  uint   `json:"curDeck"`
	NewSpread    uint   `json:"newSpread"`
	CollapseTime uint   `json:"collapseTime"`
	TimeLimit    uint   `json:"timeLim"`
	EstTimes     bool   `json:"estTimes"`
	DueCounts    bool   `json:"dueCounts"`
	// 	NextPos       uint   `json:"nextPos"`
	// 	CurrentModel  uint   `json:"curModel"`
	SortType      string `json:"sortType"`
	SortBackwards bool   `json:"sortBackwards"`
}

type Model struct {
	Name         string
	Id           uint64
	Tags         []string
	DeckId       uint64
	Fields       []Field
	SortField    uint8
	Templates    []Template
	Type         uint8
	LatexPre     string
	LatexPost    string
	CSS          string
	ModifiedTime time.Time
}

type rawModel struct {
	Name      string     `json:"name"`
	Id        string     `json:"id"`
	Tags      []string   `json:"tags"`
	DeckId    uint64     `json:"did"`
	Fields    []Field    `json:"flds"`
	SortField uint8      `json:"sortf"`
	Templates []Template `json:"tmpls"`
	Type      uint8      `json:"type"`
	LatexPre  string     `json:"latexPre"`
	LatexPost string     `json:"latexPost"`
	CSS       string     `json:"css"`
	// Req string `json:"req"`
	Mod int64 `json:"mod"`
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

func SqliteToCollection(row map[string]interface{}) (*Collection, error) {
	c := &Collection{
		Created:        time.Unix(int64(row["crt"].(float64)), 0),
		Modified:       time.Unix(int64(row["mod"].(float64)), 0),
		SchemaModified: time.Unix(int64(row["scm"].(float64)), 0),
		Ver:            int(row["ver"].(float64)),
		LastSync:       time.Unix(int64(row["ls"].(float64)), 0),
	}
	if err := json.Unmarshal([]byte(row["conf"].(string)), &(c.Config)); err != nil {
		return nil, err
	}
	var rawModels map[string]rawModel
	if err := json.Unmarshal([]byte(row["models"].(string)), &rawModels); err != nil {
		return nil, err
	} else {
		for _, m := range rawModels {
			id, err := strconv.ParseUint(m.Id, 10, 64)
			if err != nil {
				return nil, err
			}
			c.Models = append(c.Models, Model{
				Name:         m.Name,
				Id:           id,
				Tags:         m.Tags,
				DeckId:       m.DeckId,
				Fields:       m.Fields,
				SortField:    m.SortField,
				Templates:    m.Templates,
				LatexPre:     m.LatexPre,
				LatexPost:    m.LatexPost,
				CSS:          m.CSS,
				ModifiedTime: time.Unix(m.Mod, 0),
			})
		}
	}
	return c, nil
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
	c.Cards = append(c.Cards, Card{
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
	c.Notes = append(c.Notes, Note{
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
	c.Revlog = append(c.Revlog, Review{
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
