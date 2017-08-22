// +build !js

package anki

func convertQuery(query interface{}) *basicQuery {
	return query.(*basicQuery)
}
