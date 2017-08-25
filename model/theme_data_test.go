package model

import (
	"encoding/json"

	fb "github.com/FlashbackSRS/flashback-model"
)

var realTheme = func() (theme *fb.Theme) {
	if err := json.Unmarshal([]byte(realThemeJSON), &theme); err != nil {
		panic(err)
	}
	return
}()

const realThemeJSON = `{"_id":"theme-ELr8cEJJOvJU4lYz-VTXhH8wLTo","_rev":"1-c1c0bab47110550ac3bba63b68fd695f","created":"2015-09-06T17:04:36.000000823Z","modified":"2016-09-11T19:01:39Z","name":"Basic-24b78","models":[{"id":0,"modelType":"anki-basic","name":"Basic-24b78","templates":["Card 1","Card 2"],"fields":[{"fieldType":3,"name":"Front"},{"fieldType":3,"name":"Back"}],"files":["!Basic-24b78.Card 1 answer.html","!Basic-24b78.Card 1 question.html","!Basic-24b78.Card 2 answer.html","!Basic-24b78.Card 2 question.html","$template.0.html"]}],"files":["$main.css"],"modelSequence":1,"type":"theme","imported":"2017-08-21T00:19:31.776470945+02:00","_attachments":{"!Basic-24b78.Card 1 answer.html":{"content_type":"text/html","revpos":1,"digest":"md5-l0+ESIPqKXLSgF9OGHt4TA==","length":72,"stub":true},"!Basic-24b78.Card 1 question.html":{"content_type":"text/html","revpos":1,"digest":"md5-w6HfyHhRnuFV8GVcLv18Tw==","length":20,"stub":true},"!Basic-24b78.Card 2 answer.html":{"content_type":"text/html","revpos":1,"digest":"md5-eczo2+L8TO2k1xbAnaLDjQ==","length":39,"stub":true},"!Basic-24b78.Card 2 question.html":{"content_type":"text/html","revpos":1,"digest":"md5-W8cBpsiOtoM7Itj2lJDphA==","length":19,"stub":true},"$main.css":{"content_type":"text/css","revpos":1,"digest":"md5-Z90gPjA4mQo3kEzeciAEZw==","length":111,"stub":true},"$template.0.html":{"content_type":"text/html","revpos":1,"digest":"md5-73ArjFxtAiEl4YNZEfcTkA==","length":436,"stub":true}}}`
