// +build !cordova

package loginhandler

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/FlashbackSRS/flashback/model"
	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"github.com/pkg/errors"
)

func setLoginHandler(container jquery.JQuery, rel, href string) {
	li := jQuery("li."+rel, container)
	li.Show()
	jQuery("a", li).SetAttr("href", href)
}

// BTCallback handles web logins.
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

func displayError(msg string) {
	log.Printf("Authentication error: %s\n", msg)
	container := jQuery(":mobile-pagecontainer")
	jQuery("#auth_fail_reason", container).SetText(msg)
	jQuery(".show-until-load", container).Hide()
	jQuery(".hide-until-load", container).Show()
}
