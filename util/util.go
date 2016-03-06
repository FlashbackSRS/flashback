package util

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"github.com/flimzy/go-pouchdb"
	"github.com/flimzy/go-pouchdb/plugins/find"
	"github.com/gopherjs/gopherjs/js"

	"github.com/flimzy/flashback/data"
)

// JqmTargetUri determines the target URI based on a jQuery Mobile event 'ui' object
func JqmTargetUri(ui *js.Object) string {
	rawUrl := ui.Get("toPage").String()
	if rawUrl == "[object Object]" {
		rawUrl = ui.Get("toPage").Call("jqmData", "url").String()
	}
	pageUrl, _ := url.Parse(rawUrl)
	pageUrl.Path = strings.TrimPrefix(pageUrl.Path, "/android_asset/www")
	pageUrl.Host = ""
	pageUrl.User = nil
	pageUrl.Scheme = ""
	return pageUrl.String()
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

type reviewDoc struct {
	Id        string `json:"_id"`
	Rev       string `json:"_rev"`
	CurrentDb string `json:"CurrentDb"`
}

func LogReview(r *data.Review) error {
	fmt.Printf("I'm storing this review now\n")
	db, err := ReviewsDb()
	if err != nil {
		return err
	}
	_, err = db.Put(r)
	if err != nil {
		fmt.Printf("Error storing review 0: %s\n", err)
		return err
	}
	return nil
}

type DbList struct {
	Id  string   `json:"_id"`
	Rev string   `json:"_rev"`
	Dbs []string `json:"$dbs"`
}

func ReviewsDb() (*pouchdb.PouchDB, error) {
	var list DbList
	db := UserDb()
	if err := db.Get("_local/ReviewsDbs", &list, pouchdb.Options{}); err != nil {
		if !pouchdb.IsNotExist(err) {
			fmt.Printf("Error fetching _local/ReivewsDbs: %s\n", err)
			return nil, err
		}
		list = DbList{
			Id:  "_local/ReviewsDbs",
			Dbs: []string{"reviews-0-" + CurrentUser()},
		}
		_, err := db.Put(list)
		if err != nil {
			fmt.Printf("ugh: %s\n", err)
			return nil, err
		}
	}
	dbName := list.Dbs[len(list.Dbs)-1]
	return pouchdb.New(dbName), nil
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

func BaseURI() string {
	return js.Global.Get("jQuery").Get("mobile").Get("path").Call("getDocumentBase").String()
}
