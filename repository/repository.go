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

	"github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/util"
)

var couchHost string

func init() {
	couchHost = util.FindLink("flashbackdb")
}

// DB provides a simple wrapper around a pouchdb.DB object, complete with
// the pouchdb-find plugin
type DB struct {
	*pouchdb.PouchDB
	*find.PouchPluginFind
	DBName string
}

type dbInitFunc func(*DB) error

var initFuncs = map[string]dbInitFunc{
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

// BundleDB returns a DB handle for the Bundle
func BundleDB(b *fb.Bundle) (*DB, error) {
	return NewDB(b.ID.String())
}

// NewRemoteDB returns a DB handle to a remote (CouchDB) instance
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

// NewDB returns a DB handle, complete with binding to the Find plugin, to the
// requested DB
func NewDB(name string) (*DB, error) {
	db := newDB(name)
	parts := strings.SplitN(name, "-", 2)
	if initFunc, ok := initFuncs[parts[0]]; ok {
		fmt.Printf("Initializing DB %s\n", name)
		return db, initFunc(db)
	}
	return db, nil
}

// User provides a wrapper around fb.User
type User struct {
	*fb.User
}

// DB returns a DB handle to a User's DB
func (u *User) DB() (*DB, error) {
	return NewDB(u.ID.String())
}

// func (u *User) MasterReviewsDBName() string {
// 	return "reviews-" + u.ID.String()
// }
//
var currentUser *User

// CurrentUser returns a User object representing the currently logged in user, if any
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

// Compact compacts the requested DB
func (db *DB) Compact() error {
	return db.PouchDB.Compact(pouchdb.Options{})
}

// FlashbackDoc is a generic interface for all types of FB docs
type FlashbackDoc interface {
	DocID() string
	SetRev(string)
	ImportedTime() *time.Time
	ModifiedTime() *time.Time
	MergeImport(interface{}) (bool, error)
	UnmarshalJSON([]byte) error
}

// Save attempts to save any valid FlashbackDoc, including merging in case the
// document already exists from a previous import.
func (db *DB) Save(doc FlashbackDoc) error {
	if rev, err := db.Put(doc); err != nil {
		if pouchdb.IsConflict(err) {
			existing := reflect.New(reflect.TypeOf(doc).Elem()).Interface().(FlashbackDoc)
			if e := db.Get(doc.DocID(), &existing, pouchdb.Options{}); e != nil {
				return e
			}
			if doc.ImportedTime() == nil {
				// Don't attempt to merge a non-import
				return err
			}
			imported := existing.ImportedTime()
			if imported == nil {
				return err
			}
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
	} else {
		doc.SetRev(rev)
	}
	return nil
}

// Save saves the requsted FlashbackDoc to the user's DB
func (u *User) Save(doc FlashbackDoc) error {
	db, err := u.DB()
	if err != nil {
		return err
	}
	return db.Save(doc)
}
