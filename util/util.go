package util

import (
	"net/url"
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

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
	return "/" + path
}