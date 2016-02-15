package anki

type Database struct {
	Collections []*Collection
	Notes       []*Note
	Cards       []*Card
	Revlog      []*Revlog
	Grave       []*Grave
}

/*
 * https://gist.github.com/sartak/3921255
 * https://github.com/ankidroid/Anki-Android/wiki/Database-Structure
 */

// CREATE TABLE col (
//     id              integer primary key,
//     crt             integer not null,
//     mod             integer not null,
//     scm             integer not null,
//     ver             integer not null,
//     dty             integer not null,
//     usn             integer not null,
//     ls              integer not null,
//     conf            text not null,
//     models          text not null,
//     decks           text not null,
//     dconf           text not null,
//     tags            text not null
// );

type Collection struct {
	Id     int
	Crt    int
	Mod    int
	Scm    int
	Ver    int
	Dty    int
	Usn    int
	Ls     int
	Conf   string
	Models string
	Decks  string
	Dconf  string
	Tags   string
}

func SqliteToCollection(row map[string]interface{}) *Collection {
	return &Collection{
		Id:     int(row["id"].(float64)),
		Crt:    int(row["crt"].(float64)),
		Mod:    int(row["mod"].(float64)),
		Scm:    int(row["scm"].(float64)),
		Ver:    int(row["ver"].(float64)),
		Dty:    int(row["dty"].(float64)),
		Usn:    int(row["usn"].(float64)),
		Ls:     int(row["ls"].(float64)),
		Conf:   row["conf"].(string),
		Models: row["models"].(string),
		Decks:  row["decks"].(string),
		Dconf:  row["dconf"].(string),
		Tags:   row["tags"].(string),
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
	Id     int
	Nid    int
	Did    int
	Ord    int
	Mod    int
	Usn    int
	Type   int
	Queue  int
	Due    int
	Ivl    int
	Factor int
	Reps   int
	Lapses int
	Left   int
	Odue   int
	Odid   int
	Flags  int
	Data   string
}

func SqliteToCard(row map[string]interface{}) *Card {
	return &Card{
		Id:     int(row["id"].(float64)),
		Nid:    int(row["nid"].(float64)),
		Did:    int(row["did"].(float64)),
		Ord:    int(row["ord"].(float64)),
		Mod:    int(row["ord"].(float64)),
		Usn:    int(row["usn"].(float64)),
		Type:   int(row["type"].(float64)),
		Queue:  int(row["queue"].(float64)),
		Due:    int(row["due"].(float64)),
		Ivl:    int(row["ivl"].(float64)),
		Factor: int(row["factor"].(float64)),
		Reps:   int(row["reps"].(float64)),
		Lapses: int(row["lapses"].(float64)),
		Left:   int(row["left"].(float64)),
		Odue:   int(row["odue"].(float64)),
		Odid:   int(row["odid"].(float64)),
		Flags:  int(row["flags"].(float64)),
		Data:   row["data"].(string),
	}
}

// CREATE TABLE graves (
//     usn             integer not null,
//     oid             integer not null,
//     type            integer not null
// );

type Grave struct {
	Usn  int
	Oid  int
	Type int
}

func SqliteToGrave(row map[string]interface{}) *Grave {
	return &Grave{
		Usn:  int(row["usn"].(float64)),
		Oid:  int(row["oid"].(float64)),
		Type: int(row["type"].(float64)),
	}
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
	Id    int
	Guid  string
	Mid   int
	Mod   int
	Usn   int
	Tags  string
	Flds  int
	Sfld  string
	Csum  int
	Flags int
	Data  string
}

func SqliteToNote(row map[string]interface{}) *Note {
	return &Note{
		Id:    int(row["id"].(float64)),
		Guid:  row["guid"].(string),
		Mid:   int(row["mid"].(float64)),
		Mod:   int(row["mod"].(float64)),
		Usn:   int(row["usn"].(float64)),
		Tags:  row["tags"].(string),
		Flds:  int(row["flags"].(float64)),
		Sfld:  row["sfld"].(string),
		Csum:  int(row["csum"].(float64)),
		Flags: int(row["flags"].(float64)),
		Data:  row["data"].(string),
	}
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

type Revlog struct {
	Id      int
	Cid     int
	Usn     int
	Ease    int
	Ivl     int
	LastIvl int
	Factor  int
	Time    int
	Type    int
}

func SqliteToRevlog(row map[string]interface{}) *Revlog {
	return &Revlog{
		Id:      int(row["id"].(float64)),
		Cid:     int(row["cid"].(float64)),
		Usn:     int(row["usn"].(float64)),
		Ease:    int(row["ease"].(float64)),
		Ivl:     int(row["ivl"].(float64)),
		LastIvl: int(row["lastIvl"].(float64)),
		Factor:  int(row["factor"].(float64)),
		Time:    int(row["time"].(float64)),
		Type:    int(row["type"].(float64)),
	}
}
