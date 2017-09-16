// +build debug

package main

import "github.com/FlashbackSRS/flashback/oauth2"

func providerInit(baseURL, facebookID string) map[string]string {
	return map[string]string{
		"facebook": oauth2.FacebookURL(facebookID, baseURL),
		"devlogin": "#devlogin",
	}
}
