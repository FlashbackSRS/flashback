package util

import (
	"encoding/base64"
	"net/url"
	"strings"

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
func GetUserFromCookie() string {
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