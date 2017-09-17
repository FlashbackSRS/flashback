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

func setLoginHandler(_ *model.Repo, container jquery.JQuery, rel, href string, cancel func()) {
	li := jQuery("li."+rel, container)
	li.Show()
	a := jQuery("a", li)
	a.SetAttr("href", href)
	a.On("click", cancel)
}

// BTCallback handles web logins.
func BTCallback(repo *model.Repo, providers map[string]string) jqeventrouter.HandlerFunc {
	devLoginHandler := devLogin(repo)
	return func(event *jquery.Event, ui *js.Object, params url.Values) bool {
		if params.Get("provider") == "devlogin" {
			log.Debug("Callback for dev login")
			return devLoginHandler(event, ui, params)
		}
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

// checkLoginStatus checks for auth in the background
func checkLoginStatus(repo *model.Repo) func() {
	log.Debug("checkLoginStatus\n")
	defer log.Debug("return from checkLoginStatus\n")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		if _, err := repo.CurrentUser(); err != nil {
			log.Debugf("(cls) repo err: %s", err)
			return
		}
		if e := checkCtx(ctx); e != nil {
			log.Debugf("(cls) ctx err: %s", e)
			return
		}

		log.Debugln("(cls) Already authenticated")
		js.Global.Get("jQuery").Get("mobile").Call("changePage", "index.html")
	}()
	return cancel
}
