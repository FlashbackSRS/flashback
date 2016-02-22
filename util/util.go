package util

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/go-pouchdb/plugins/find"
	"github.com/gopherjs/gopherjs/js"
)

// JqmTargetUri determines the target URI based on a jQuery Mobile event 'ui' object
func JqmTargetUri(ui *js.Object) string {
	rawUrl := ui.Get("toPage").String()
	if rawUrl == "[object Object]" {
		rawUrl = ui.Get("toPage").Call("jqmData", "url").String()
	}
	pageUrl, _ := url.Parse(rawUrl)
	path := strings.TrimPrefix(pageUrl.Path, "/android_asset/www")
	if len(pageUrl.Fragment) > 0 {
		return path + "#" + pageUrl.Fragment
	}
	return "/" + strings.TrimPrefix(path, "/")
}

// UserFromCookie extracts a user name from the CouchDB cookie, which is set
// during the authentication phase
func CurrentUser() string {
	value := GetCouchCookie(js.Global.Get("document").Get("cookie").String())
	userid := ExtractUserID(value)
	return userid
}

func GetCouchCookie(cookieHeader string) string {
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

func ExtractUserID(cookieValue string) string {
	decoded, _ := base64.StdEncoding.DecodeString(cookieValue)
	values := strings.Split(string(decoded), ":")
	return values[0]
}

func CouchHost() string {
	return findLink("flashbackdb")
}

func FlashbackHost() string {
	return findLink("flashback")
}

func findLink(rel string) string {
	links := js.Global.Get("document").Call("getElementsByTagName", "head").Index(0).Call("getElementsByTagName", "link")
	for i := 0; i < links.Length(); i++ {
		link := links.Index(i)
		if link.Get("rel").String() == rel {
			return link.Get("href").String()
		}
	}
	return ""
}

func UserDb() *pouchdb.PouchDB {
	dbName := "user-" + CurrentUser()
	return pouchdb.New(dbName)
}

var initMap = make(map[string]<-chan struct{})

// InitUserDb will do any db initialization necessary after account
// creation/login such as setting up indexes. The return value is a channel,
// which will be closed when initialization is complete
func InitUserDb() <-chan struct{} {
	user := CurrentUser()
	if done, ok := initMap[user]; ok {
		// Initialization has already been started, so return the existing
		// channel (which may already be closed)
		return done
	}
	done := make(chan struct{})
	initMap[user] = done
	go func() {
		db := UserDb()
		dbFind := find.New(db)
		fmt.Printf("Creating index...\n")
		err := dbFind.CreateIndex(find.Index{
			Name:   "type",
			Fields: []string{"$type"},
		})
		if err != nil && !find.IsIndexExists(err) {
			fmt.Printf("Error creating index: %s\n", err)
		}
		close(done)
	}()
	return done
}
