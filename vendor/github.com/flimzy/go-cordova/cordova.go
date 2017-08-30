package cordova

import (
	"strings"
	"github.com/gopherjs/gopherjs/js"
)

func global() *js.Object {
	if m := js.Global.Get("cordova"); m != nil {
		return m
	}
	if m := js.Global.Get("PhoneGap"); m != nil {
		return m
	}
	if m := js.Global.Get("phonegap"); m != nil {
		return m
	}
	return nil
}

func Global() *js.Object {
	mobile := global()
	if mobile == nil {
		return nil
	}
	ua := strings.ToLower(js.Global.Get("navigator").Get("userAgent").String())

	if strings.HasPrefix(strings.ToLower(js.Global.Get("location").Get("href").String()), "file:///") &&
		(strings.Contains(ua, "ios") || strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "android")) {
		return mobile
	}
	return nil
}


func IsMobile() bool {
	if m := Global(); m == nil {
		return false
	}
	return true
}
