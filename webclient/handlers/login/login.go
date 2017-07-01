// +build js

package loginhandler

import (
	"fmt"
	"net/url"

	"github.com/flimzy/go-cordova"
	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/pkg/errors"
	"honnef.co/go/js/console"

	"github.com/FlashbackSRS/flashback/config"
	"github.com/FlashbackSRS/flashback/repository"
)

var jQuery = jquery.NewJQuery

const facebookAuthURL = "https://www.facebook.com/v2.9/dialog/oauth"

func facebookURL(conf *config.Conf) string {
	redirParams := url.Values{}
	redirParams.Add("provider", "facebook")
	redir, err := url.Parse(conf.GetString("flashback_app"))
	if err != nil {
		panic("Invalid flashback_app URL: " + err.Error())
	}
	redir.RawQuery = redirParams.Encode()

	params := url.Values{}
	params.Add("client_id", conf.GetString("facebook_client_id"))
	params.Add("redirect_uri", redir.String())
	params.Add("response_type", "token")
	return fmt.Sprintf("%s?%s", facebookAuthURL, params.Encode())
}

// BeforeTransition prepares the logout page before display.
func BeforeTransition(conf *config.Conf) jqeventrouter.HandlerFunc {
	providers := map[string]string{
		"facebook": facebookURL(conf),
	}

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
		console.Log("Success logging in")
		u, err := repo.CurrentUser()
		if err != nil {
			log.Debugf("No user logged in?? %s\n", err)
		} else {
			// To make sure the DB is initialized as soon as possible
			u.DB()
		}
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

func BTCallback(conf *config.Conf) jqeventrouter.HandlerFunc {
	return func(_ *jquery.Event, ui *js.Object, _ url.Values) bool {
		log.Debug("Auth Callback")
		provider, token, err := extractAuthToken(js.Global.Get("location").String())
		if err != nil {
			displayError(err.Error())
			return true
		}
		go func() {
			container := jQuery(":mobile-pagecontainer")
			if err := repo.Auth(provider, token, conf.GetString("flashback_api")); err != nil {
				displayError(err.Error())
			} else {
				ui.Set("toPage", "index.html")
				container.Trigger("pagecontainerbeforechange", ui)
			}
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
