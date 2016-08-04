package util

import (
	"encoding/base64"
	// 	"errors"
	// 	"fmt"
	"net/url"
	"strings"

	// 	"github.com/flimzy/go-pouchdb"
	"github.com/gopherjs/gopherjs/js"
	// 	"github.com/flimzy/flashback-model"
	// 	"github.com/flimzy/flashback/repository"
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

func FlashbackHost() string {
	return FindLink("flashback")
}

func FindLink(rel string) string {
	links := js.Global.Get("document").Call("getElementsByTagName", "head").Index(0).Call("getElementsByTagName", "link")
	for i := 0; i < links.Length(); i++ {
		link := links.Index(i)
		if link.Get("rel").String() == rel {
			return link.Get("href").String()
		}
	}
	return ""
}

/*
func UserDb() (*pouchdb.PouchDB, error) {
	userName := CurrentUser()
	if userName == "" {
		return nil, errors.New("Not logged in")
	}
	return pouchdb.New("user-" + userName), nil
}
*/

type reviewDoc struct {
	Id        string `json:"_id"`
	Rev       string `json:"_rev"`
	CurrentDb string `json:"CurrentDb"`
}

/*
func LogReview(r *fb.Review) error {
	db, err := ReviewsDb()
	if err != nil {
		return err
	}
	_, err = db.Put(r)
	if err != nil {
		return err
	}
	return nil
}

type DbList struct {
	Id  string   `json:"_id"`
	Rev string   `json:"_rev"`
	Dbs []string `json:"$dbs"`
}

func getReviewsDbList() (DbList, error) {
	var list DbList
	db, err := UserDb()
	if err != nil {
		return list, err
	}
	if err := db.Get("_local/ReviewsDbs", &list, pouchdb.Options{}); err != nil && pouchdb.IsNotExist(err) {
		return DbList{Id: "_local/ReviewsDbs"}, nil
	}
	return list, err
}

func setReviewsDbList(list DbList) error {
	if list.Id != "_local/ReviewsDbs" {
		return fmt.Errorf("Invalid id '%s' for ReviewsDbs", list.Id)
	}
	fmt.Printf("Setting list to: %v\n", list.Dbs)
	db, err := UserDb()
	if err != nil {
		return err
	}
	_, err = db.Put(list)
	return err
}

func ReviewsSyncDbs() (*repo.DB, error) {
	userName := CurrentUser()
	if userName == "" {
		return nil, errors.New("Not logged in")
	}
	list, err := getReviewsDbList()
	if err != nil {
		return nil, err
	}
	if len(list.Dbs) == 0 {
		// No reviews database, nothing to sync
		return nil, nil
	}
	dbName := list.Dbs[0]
	db := repo.NewDB(dbName)
	if len(list.Dbs) > 1 {
		fmt.Printf("WARNING: More than one active reviews database!\n")
		return db, nil
	}
	var newDbPrefix string = "reviews-1-"
	if strings.HasPrefix(dbName, "reviews-1-") {
		newDbPrefix = "reviews-0-"
	}
	list.Dbs = append(list.Dbs, newDbPrefix+userName)
	if err := setReviewsDbList(list); err != nil {
		return nil, err
	}
	return db, nil
}

func ZapReviewsDb(db *repo.DB) error {
	info, err := db.Info()
	if err != nil {
		return err
	}
	list, err := getReviewsDbList()
	if err != nil {
		return err
	}
	if list.Dbs[0] != info.DBName {
		return fmt.Errorf("Attempt to remove ReviewsDb '%s' not at head of list", info.DBName)
	}
	list.Dbs = list.Dbs[1:]
	if err := setReviewsDbList(list); err != nil {
		return err
	}
	return db.Destroy(pouchdb.Options{})
}

func ReviewsDb() (*pouchdb.PouchDB, error) {
	list, err := getReviewsDbList()
	if err != nil {
		return nil, err
	}
	if len(list.Dbs) == 0 {
		list.Dbs = []string{"reviews-0-" + CurrentUser()}
		err := setReviewsDbList(list)
		if err != nil {
			return nil, err
		}
	}
	dbName := list.Dbs[len(list.Dbs)-1]
	return pouchdb.New(dbName), nil
}
*/
var initMap = make(map[string]<-chan struct{})

func BaseURI() string {
	rawUri := js.Global.Get("jQuery").Get("mobile").Get("path").Call("getDocumentBase").String()
	uri, _ := url.Parse(rawUri)
	uri.Fragment = ""
	uri.RawQuery = ""
	return uri.String()
}
