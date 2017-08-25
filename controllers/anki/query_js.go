// +build js

package anki

import (
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

func convertQuery(query interface{}) *basicQuery {
	if jsQuery, ok := query.(*js.Object); ok {
		typedAnswers := make(map[string]string)
		for _, k := range js.Keys(jsQuery) {
			if strings.HasPrefix(k, typePrefix) {
				typedAnswers[strings.TrimPrefix(k, typePrefix)] = jsQuery.Get(k).String()
			}
		}
		return &basicQuery{
			Submit:       jsQuery.Get("submit").String(),
			TypedAnswers: typedAnswers,
		}
	}
	return query.(*basicQuery)
}
