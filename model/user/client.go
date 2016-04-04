// +build js

package user

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/pborman/uuid"

	"github.com/flimzy/go-pouchdb/plugins/find"
	"github.com/gopherjs/gopherjs/js"

	"github.com/flimzy/flashback/model"
)

var currentUser *User

func CurrentUser() (*User, error) {
	cookie := getCouchCookie(js.Global.Get("document").Get("cookie").String())
	userid := extractUserID(cookie)
	id := uuid.Parse(userid)
	if currentUser != nil && uuid.Equal(currentUser.uuid, id) {
		return currentUser, nil
	}
	u := &User{
		uuid: id,
	}
	u.newUserDoc()
	// TODO: Populate user object with values from DB
	// 	u, err := Fetch( id )
	// 	if err != nil {
	// 		return nil, err
	// 	}
	currentUser = u
	return currentUser, nil
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

func extractUserID(cookieValue string) string {
	decoded, _ := base64.StdEncoding.DecodeString(cookieValue)
	values := strings.Split(string(decoded), ":")
	return values[0]
}

func (u *User) UserDB() *model.DB {
	return model.NewDB(u.DBName())
}

// func Fetch(id uuid.UUID) (*User, error) {
// 	u := &User{}
// 	db := pouchdb.New("flashback")
// 	err := db.Get(id.String(), u, pouchdb.Options{})
// 	if err != nil && pouchdb.IsNotExist(err) {
// 		db := pouchdb.New( util.CouchHost() + "/_users"
// 	}
// 	return u, err
//
// }

// InitDB will do any db initialization necessary after account
// creation/login such as setting up indexes. The return value is a channel,
// which will be closed when initialization is complete
func (u *User) InitDB() <-chan struct{} {
	if u.dbInitDone != nil {
		return u.dbInitDone
	}
	done := make(chan struct{})
	u.dbInitDone = done
	go func() {
		db := u.UserDB()
		fmt.Printf("Creating index...\n")
		err := db.CreateIndex(find.Index{
			Name:   "type",
			Fields: []string{"type"},
		})
		if err != nil && !find.IsIndexExists(err) {
			fmt.Printf("Error creating index: %s\n", err)
		}
		close(done)
	}()
	return u.dbInitDone
}
