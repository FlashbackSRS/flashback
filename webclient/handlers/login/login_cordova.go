// +build cordova

package loginhandler

import (
	"context"
	"fmt"
	"sync"

	"github.com/FlashbackSRS/flashback/model"
	"github.com/flimzy/jqeventrouter"
	"github.com/flimzy/log"
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jquery"
	"honnef.co/go/js/console"
)

func setLoginHandler(repo *model.Repo, container jquery.JQuery, rel, href string, cancel func()) {
	li := jQuery("li."+rel, container)
	li.Show()
	jQuery("a", li).On("click", func() {
		cancel()
		CordovaLogin(repo)
	})
}

// CordovaLogin handles login for the Cordova runtime.
func CordovaLogin(repo *model.Repo) bool {
	console.Log("CordovaLogin()")
	js.Global.Get("facebookConnectPlugin").Call("login", []string{}, func(response *js.Object) {
		log.Debug("cl success pre goroutine\n")
		go func() {
			provider := "facebook"
			token := response.Get("authResponse").Get("accessToken").String()
			if err := repo.Auth(context.TODO(), provider, token); err != nil {
				displayError(err.Error())
				return
			}
			fmt.Printf("Auth succeeded!\n")
			js.Global.Get("jQuery").Get("mobile").Call("changePage", "index.html")
		}()
	}, func(e *js.Object) {
		log.Printf("Failure logging in: %s", e.Get("errorMessage").String())
	})
	console.Log("Leaving CordovaLogin()")
	return false
}

// BTCallback defers to devLogin in Cordova.
func BTCallback(repo *model.Repo, _ map[string]string) jqeventrouter.HandlerFunc {
	return devLogin(repo)
}

// checkLoginStatus checks for auth in the background
func checkLoginStatus(repo *model.Repo) func() {
	log.Debug("checkLoginStatus\n")
	defer log.Debug("return from checkLoginStatus\n")
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer cancel()
		var wg sync.WaitGroup
		wg.Add(1)
		js.Global.Get("facebookConnectPlugin").Call("getLoginStatus", func(response *js.Object) {
			go func() {
				defer wg.Done()
				provider := "facebook"
				if authResponse := response.Get("authResponse"); authResponse != js.Undefined {
					token := authResponse.Get("accessToken").String()
					fmt.Printf("token = %s\n", token)
					if err := repo.Auth(context.TODO(), provider, token); err != nil {
						log.Printf("(cls) Auth error: %s", err)
						return
					}
				}
			}()
		}, func() {
			wg.Done()
		})
		wg.Wait()
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
