package anki

import (
	"time"
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
//     usn             integer not null,
//     ls              integer not null,
//     conf            text not null,
//     models          text not null,
//     decks           text not null,
//     dconf           text not null,
//     tags            text not null
// );

type Collection struct {
	Created        time.Time
	Modified       time.Time
	SchemaModified time.Time
	Ver            int
	Usn      int
	LastSync time.Time
	Conf     string
	Models   string
	Decks    string
	Dconf    string
	Tags     string
	Cards    []Card
	Notes    []Note
	Revlog   []Review
}

func SqliteToCollection(row map[string]interface{}) *Collection {
	return &Collection{
		Created:        time.Unix(int64(row["crt"].(float64)), 0),
		Modified:       time.Unix(int64(row["mod"].(float64)), 0),
		SchemaModified: time.Unix(int64(row["scm"].(float64)), 0),
		Ver:            int(row["ver"].(float64)),
		Usn:      int(row["usn"].(float64)),
		LastSync: time.Unix(int64(row["ls"].(float64)), 0),
		Conf:     row["conf"].(string),
		Models:   row["models"].(string),
		Decks:    row["decks"].(string),
		Dconf:    row["dconf"].(string),
		Tags:     row["tags"].(string),
	}
}

// CREATE TABLE cards (
//     id              integer primary key,   /* 0 */
//     nid             integer not null,      /* 1 */
//     did             integer not null,      /* 2 */
//     ord             integer not null,      /* 3 */
//     mod             integer not null,      /* 4 */
//     usn             integer not null,      /* 5 */
//     type            integer not null,      /* 6 */
//     queue           integer not null,      /* 7 */
//     due             integer not null,      /* 8 */
//     ivl             integer not null,      /* 9 */
//     factor          integer not null,      /* 10 */
//     reps            integer not null,      /* 11 */
//     lapses          integer not null,      /* 12 */
//     left            integer not null,      /* 13 */
//     odue            integer not null,      /* 14 */
//     odid            integer not null,      /* 15 */
//     flags           integer not null,      /* 16 */
//     data            text not null          /* 17 */
// );
// CREATE INDEX ix_cards_usn on cards (usn);
// CREATE INDEX ix_cards_nid on cards (nid);
// CREATE INDEX ix_cards_sched on cards (did, queue, due);

type Card struct {
	Id       uint64
	Nid      uint64
	Did      uint64
	Ord      int
	Modified time.Time
	Usn      int
	Type     CardType
	Queue    QueueType
	Due      int
	Interval int
	Factor   int
	Reps     int
	Lapses   int
	Left     int
	// 	Odue     int
	// 	Odid     int
	// 	Flags    int
	Data string
}

func (c *Collection) AddCardFromSqlite(row map[string]interface{}) {
	c.Cards = append(c.Cards, Card{
		Id:       uint64(row["id"].(float64)),
		Nid:      uint64(row["nid"].(float64)),
		Did:      uint64(row["did"].(float64)),
		Ord:      int(row["ord"].(float64)),
		Modified: time.Unix(int64(row["ord"].(float64)), 0),
		Usn:      int(row["usn"].(float64)),
		Type:     CardType(row["type"].(float64)),
		Queue:    QueueType(row["queue"].(float64)),
		Due:      int(row["due"].(float64)),
		Interval: int(row["ivl"].(float64)),
		Factor:   int(row["factor"].(float64)),
		Reps:     int(row["reps"].(float64)),
		Lapses:   int(row["lapses"].(float64)),
		Left:     int(row["left"].(float64)),
		// 		Odue:     int(row["odue"].(float64)),
		// 		Odid:     int(row["odid"].(float64)),
		// 		Flags:    int(row["flags"].(float64)),
		Data: row["data"].(string),
	})
}

// CREATE TABLE notes (
//     id              integer primary key,   /* 0 */
//     guid            text not null,         /* 1 */
//     mid             integer not null,      /* 2 */
//     mod             integer not null,      /* 3 */
//     usn             integer not null,      /* 4 */
//     tags            text not null,         /* 5 */
//     flds            text not null,         /* 6 */
//     sfld            integer not null,      /* 7 */
//     csum            integer not null,      /* 8 */
//     flags           integer not null,      /* 9 */
//     data            text not null          /* 10 */
// );
// CREATE INDEX ix_notes_usn on notes (usn);
// CREATE INDEX ix_notes_csum on notes (csum);

type Note struct {
	Id       uint64
	Guid     string
	Mid      uint64
	Modified time.Time
	Usn      int
	Tags     string
	Flds     int
	Sfld     string
	Csum     uint64
	// 	Flags    int
	// 	Data     string
}

func (c *Collection) AddNoteFromSqlite(row map[string]interface{}) {
	c.Notes = append( c.Notes, Note{
		Id:       uint64(row["id"].(float64)),
		Guid:     row["guid"].(string),
		Mid:      uint64(row["mid"].(float64)),
		Modified: time.Unix(int64(row["mod"].(float64)), 0),
		Usn:      int(row["usn"].(float64)),
		Tags:     row["tags"].(string),
		Flds:     int(row["flags"].(float64)),
		Sfld:     row["sfld"].(string),
		Csum:     uint64(row["csum"].(float64)),
		// 		Flags:    int(row["flags"].(float64)),
		// 		Data:     row["data"].(string),
	})
}

// CREATE TABLE revlog (
//     id              integer primary key,
//     cid             integer not null,
//     usn             integer not null,
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
	Cid          uint64
	Usn          int
	Ease         Ease
	Interval     int
	LastInterval int
	Factor       int
	Time         time.Duration
	Type         int
}

func (c *Collection) AddReviewFromSqlite(row map[string]interface{}) {
	c.Revlog = append( c.Revlog, Review{
		Id:           uint(row["id"].(float64)),
		Cid:          uint64(row["cid"].(float64)),
		Usn:          int(row["usn"].(float64)),
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
