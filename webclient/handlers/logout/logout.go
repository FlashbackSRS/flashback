package logout

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
)

var jQuery = jquery.NewJQuery

func BeforeTransition(event *jquery.Event, ui *js.Object) bool {
	fmt.Printf("logout BEFORE\n")

	button := jQuery("#logout")

	button.On("click", func() {
		fmt.Printf("Trying to log out now\n")
		expireTime, _ := time.Parse(time.RFC3339, "1900-01-01T00:00:00+00:00")
		emptyCookie := &http.Cookie{
			Name:    "AuthSession",
			Expires: expireTime,
		}
		js.Global.Get("document").Set("cookie", emptyCookie.String())
		jQuery(":mobile-pagecontainer").Call("pagecontainer", "change", "/")
	})
	return true
}
