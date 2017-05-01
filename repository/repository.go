package repo

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/pouchdb" // PouchDB driver
	"github.com/gopherjs/gopherjs/js"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"

	"github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/util"
)

// PouchDBOptions is passed to Open(), and exists for the sake of automated
// tests. It should generally be ignored otherwise.
var PouchDBOptions kivik.Options

var couchHost string

func init() {
	couchHost = util.FindLink("flashbackdb")
}

// DB provides a simple wrapper around a DB object.
type DB struct {
	*kivik.DB
	User   *User
	DBName string
}

type dbInitFunc func(*DB) error

var initFuncs = map[string]dbInitFunc{
	"user":   userDBInit,
	"bundle": bundleDBInit,
}

// BundleDB returns a DB handle for the Bundle
func (u *User) BundleDB(b *fb.Bundle) (*DB, error) {
	return u.NewDB(b.ID.String())
}

// NewRemoteDB returns a DB handle to a remote (CouchDB) instance
func (u *User) NewRemoteDB(name string) (*DB, error) {
	return u.newDB(couchHost + "/" + name), nil
}

func (u *User) newDB(name string) *DB {
	client, err := kivik.New(context.TODO(), "pouch", "")
	if err != nil {
		panic(err)
	}
	db, err := client.DB(context.TODO(), name, PouchDBOptions)
	if err != nil {
		panic(err)
	}
	return &DB{
		DB:     db,
		DBName: name,
		User:   u,
	}
}

// NewDB returns a DB handle, complete with binding to the Find plugin, to the
// requested DB
func (u *User) NewDB(name string) (*DB, error) {
	parts := strings.SplitN(name, "-", 2)
	if parts[0] != "user" {
		name = fmt.Sprintf("%s-%s-%s", parts[0], u.ID, parts[1])
	}
	db := u.newDB(name)
	return db, initDB(parts[0], db)
}

func initDB(name string, db *DB) error {
	if initFunc, ok := initFuncs[name]; ok {
		return errors.Wrap(initFunc(db), "db init")
	}
	return nil
}

// User provides a wrapper around fb.User
type User struct {
	*fb.User
}

// DB returns a DB handle to a User's DB
func (u *User) DB() (*DB, error) {
	return u.NewDB(u.ID.String())
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
func (db *DB) Compact(ctx context.Context) error {
	var errs error
	if err := db.DB.Compact(ctx); err != nil {
		errs = multierror.Append(errs, err)
	}
	if err := db.DB.ViewCleanup(ctx); err != nil {
		errs = multierror.Append(errs, err)
	}
	return errs
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
	var rev string
	var err error
	if rev, err = db.Put(context.TODO(), doc.DocID(), doc); err != nil {
		if kivik.StatusCode(err) != kivik.StatusConflict {
			return err
		}
		existing := reflect.New(reflect.TypeOf(doc).Elem()).Interface().(FlashbackDoc)
		row, err := db.Get(context.TODO(), doc.DocID())
		if err != nil {
			return errors.Wrap(err, "failed to fetch existing document")
		}
		if err = row.ScanDoc(&existing); err != nil {
			return errors.Wrap(err, "failed to parse existing document")
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
			// The existing document was modified after import, so we won't allow re-importing
			return err
		}
		var changed bool
		if changed, err = doc.MergeImport(existing); err != nil {
			return errors.Wrap(err, "failed to merge into existing document")
		}
		if changed {
			if rev, err = db.Put(context.TODO(), doc.DocID(), doc); err != nil {
				return errors.Wrap(err, "failed to store updated document")
			}
		}
	}
	doc.SetRev(rev)
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
