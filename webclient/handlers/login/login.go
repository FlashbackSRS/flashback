// +build js

package loginhandler

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/flimzy/go-cordova"
	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/pkg/errors"
	"honnef.co/go/js/console"

	"github.com/FlashbackSRS/flashback/model"
)

var jQuery = jquery.NewJQuery

// BeforeTransition prepares the logout page before display.
func BeforeTransition(providers map[string]string) jqeventrouter.HandlerFunc {
	return func(_ *jquery.Event, _ *js.Object, _ url.Values) bool {
		console.Log("login BEFORE")

		container := jQuery(":mobile-pagecontainer")
		for rel, href := range providers {
			li := jQuery("li."+rel, container)
			li.Show()
			a := jQuery("a", li)
			if cordova.IsMobile() {
				console.Log("Setting on click event")
				a.On("click", CordovaLogin)
			} else {
				a.SetAttr("href", href)
			}
		}
		jQuery(".show-until-load", container).Hide()
		jQuery(".hide-until-load", container).Show()

		return true
	}
}

func CordovaLogin() bool {
	console.Log("CordovaLogin()")
	js.Global.Get("facebookConnectPlugin").Call("login", []string{}, func() {
		panic("CordovaLogin needs to be made to work")
		// console.Log("Success logging in")
		// u, err := repo.CurrentUser()
		// if err != nil {
		// 	log.Debugf("No user logged in?? %s\n", err)
		// } else {
		// 	// To make sure the DB is initialized as soon as possible
		// 	u.DB()
		// }
	}, func() {
		console.Log("Failure logging in")
	})
	console.Log("Leaving CordovaLogin()")
	return false
}

func displayError(msg string) {
	log.Printf("Authentication error: %s\n", msg)
	container := jQuery(":mobile-pagecontainer")
	jQuery("#auth_fail_reason", container).SetText(msg)
	jQuery(".show-until-load", container).Hide()
	jQuery(".hide-until-load", container).Show()
}

func BTCallback(repo *model.Repo, providers map[string]string) jqeventrouter.HandlerFunc {
	return func(event *jquery.Event, ui *js.Object, _ url.Values) bool {
		log.Debug("Auth Callback")
		provider, token, err := extractAuthToken(js.Global.Get("location").String())
		if err != nil {
			displayError(err.Error())
			return true
		}
		go func() {
			if err := repo.Auth(context.TODO(), provider, token); err != nil {
				msg := err.Error()
				if strings.Contains(msg, "Session has expired on") {
					for name, href := range providers {
						if name == provider {
							log.Debugf("Redirecting unauthenticated user to %s\n", href)
							js.Global.Get("location").Call("replace", href)
							event.StopImmediatePropagation()
							return
						}
					}
				}
				displayError(msg)
				return
			}
			fmt.Printf("Auth succeeded!\n")
			ui.Set("toPage", "index.html")
			event.StopImmediatePropagation()
			js.Global.Get("jQuery").Get("mobile").Call("changePage", "index.html")
			// container.Trigger("pagecontainerbeforechange", ui)
		}()
		return true
	}
}

func extractAuthToken(uri string) (provider, token string, err error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return "", "", err
	}
	provider = parsed.Query().Get("provider")
	if provider == "" {
		return "", "", errors.New("no provider")
	}
	switch provider {
	case "facebook":
		frag, err := url.ParseQuery(parsed.Fragment)
		if err != nil {
			return "", "", errors.Wrapf(err, "failed to parse URL fragment")
		}
		token = frag.Get("access_token")
	default:
		return "", "", errors.Errorf("Unknown provider '%s'", provider)
	}
	return provider, token, nil
}
