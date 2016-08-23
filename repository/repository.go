package repo

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/pborman/uuid"

	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/go-pouchdb/plugins/find"
	"github.com/gopherjs/gopherjs/js"

	"github.com/flimzy/flashback-model"
	"github.com/flimzy/flashback/util"
)

var couchHost string

func init() {
	couchHost = util.FindLink("flashbackdb")
}

type DB struct {
	*pouchdb.PouchDB
	*find.PouchPluginFind
	DBName string
}

type dbInitFunc func(*DB) error

var initFuncs map[string]dbInitFunc = map[string]dbInitFunc{
	"user": func(db *DB) error {
		return commonDBInit(db)
	},
	"bundle": func(db *DB) error {
		return commonDBInit(db)
	},
}

func commonDBInit(db *DB) error {
	err := db.CreateIndex(find.Index{
		Fields: []string{"type"},
	})
	if err != nil && !find.IsIndexExists(err) {
		return err
	}
	return nil
}

func BundleDB(b *fb.Bundle) (*DB, error) {
	return NewDB(b.ID.String())
}

func NewRemoteDB(name string) (*DB, error) {
	return newDB(couchHost + "/" + name), nil
}

func newDB(name string) *DB {
	pdb := pouchdb.New(name)
	return &DB{
		PouchDB:         pdb,
		PouchPluginFind: find.New(pdb),
		DBName:          name,
	}
}

func NewDB(name string) (*DB, error) {
	db := newDB(name)
	parts := strings.SplitN(name, "-", 2)
	if initFunc, ok := initFuncs[parts[0]]; ok {
		fmt.Printf("Initializing DB %s\n", name)
		return db, initFunc(db)
	}
	return db, nil
}

type User struct {
	*fb.User
}

func (u *User) DB() (*DB, error) {
	return NewDB(u.ID.String())
}

func (u *User) MasterReviewsDBName() string {
	return "reviews-" + u.ID.String()
}

var currentUser *User

func CurrentUser() (*User, error) {
	cookie := getCouchCookie(js.Global.Get("document").Get("cookie").String())
	userid := extractUserID(cookie)
	id := uuid.Parse(userid)
	if id == nil {
		return nil, errors.New("Invalid user ID found in cookie")
	}
	if currentUser != nil && currentUser.Equal(id) {
		return currentUser, nil
	}
	u, err := fb.NewUser(id, "")
	if err != nil {
		return nil, err
	}
	currentUser = &User{u}
	return currentUser, nil
}

func extractUserID(cookieValue string) string {
	decoded, _ := base64.StdEncoding.DecodeString(cookieValue)
	values := strings.Split(string(decoded), ":")
	return values[0]
}

func getCouchCookie(cookieHeader string) string {
	cookies := strings.Split(cookieHeader, ";")
	for _, cookie := range cookies {
		nv := strings.Split(strings.TrimSpace(cookie), "=")
		if nv[0] == "AuthSession" {
			value, _ := url.QueryUnescape(nv[1])
			return value
		}
	}
	return ""
}

func (db *DB) Compact() error {
	return db.PouchDB.Compact(nil)
}

type FlashbackDoc interface {
	DocID() string
	SetRev(string)
	ImportedTime() *time.Time
	ModifiedTime() *time.Time
	MergeImport(interface{}) (bool, error)
	UnmarshalJSON([]byte) error
}

func (db *DB) Save(doc FlashbackDoc) error {
	if rev, err := db.Put(doc); err != nil {
		if pouchdb.IsConflict(err) {
			existing := reflect.New(reflect.TypeOf(doc).Elem()).Interface().(FlashbackDoc)
			if err := db.Get(doc.DocID(), &existing, pouchdb.Options{}); err != nil {
				return err
			}
			if doc.ImportedTime() == nil {
				// Don't attempt to merge a non-import
				return err
			}
			if imported := existing.ImportedTime(); imported == nil {
				return err
			} else {
				if existing.ModifiedTime().After(*imported) {
					// The existing document was mosified after import, so we won't allow further importing
					return err
				}
				if changed, err := doc.MergeImport(existing); err != nil {
					return err
				} else if changed {
					if rev, err = db.Put(doc); err != nil {
						return err
					}
				}
			}
		}
	} else {
		doc.SetRev(rev)
	}
	return nil
}

func (u *User) Save(doc FlashbackDoc) error {
	db, err := u.DB()
	if err != nil {
		return err
	}
	return db.Save(doc)
}
