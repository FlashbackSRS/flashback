// +build js

package all_pages

import (
	"encoding/base64"
	"strings"
	"net/url"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"golang.org/x/net/context"
	"honnef.co/go/js/console"

	"github.com/flimzy/flashback/webclient/pages"
	"github.com/flimzy/flashback/clientstate"
)

func BeforeChange(ctx context.Context, event *jquery.Event, ui *js.Object) pages.Action {
	console.Log("ALL BEFORE")
	state := ctx.Value("AppState").(*clientstate.State)
	target := ctx.Value("target").(string)

	if state.CurrentUser == "" {
		state.CurrentUser = UserFromCookie()
	}

	if target != "login.html" {
		if state.CurrentUser == "" {
			console.Log("No local user defined")
			return pages.Redirect("login.html")
		}
	}
	return pages.Return()
}

func UserFromCookie() string {
	value := GetCouchCookie( js.Global.Get("document").Get("cookie").String() )
	userid := ExtractUserID( value )
	return userid
}

func GetCouchCookie(cookieHeader string) string {
	cookies := strings.Split( cookieHeader, ";" )
	for _,cookie := range cookies {
		nv := strings.Split( strings.TrimSpace(cookie), "=")
		if nv[0] == "AuthSession" {
			value,_ := url.QueryUnescape( nv[1] )
			return value
		}
	}
	return ""
}

func ExtractUserID(cookieValue string) string {
	decoded,_ := base64.StdEncoding.DecodeString(cookieValue)
	values := strings.Split(string(decoded), ":")
	return values[0]
}
