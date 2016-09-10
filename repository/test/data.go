package test

import (
	"encoding/json"
	"fmt"
)

var expected = map[string][]byte{
	"theme-uQ3TFsQgm9Y29vlgC-lphauhK3M": []byte(`{
		"_attachments": {
			"!Basic-24b78.Card 1 answer.html": {
				"content_type": "text/html",
				"digest": "md5-FNeLYevm/jH+BOA9IVnZPQ==",
				"length": 75,
				"revpos": 1,
				"stub": true
			},
			"!Basic-24b78.Card 1 question.html": {
				"content_type": "text/html",
				"digest": "md5-fOa6+X5wsJ+1Vt5Q0v5VxA==",
				"length": 10,
				"revpos": 1,
				"stub": true
			},
			"!Basic-24b78.Card 2 answer.html": {
				"content_type": "text/html",
				"digest": "md5-6IlntfutXQz3EVecHFYlnw==",
				"length": 31,
				"revpos": 1,
				"stub": true
			},
			"!Basic-24b78.Card 2 question.html": {
				"content_type": "text/html",
				"digest": "md5-E51mS9DsSfBGiL3sXUUUVA==",
				"length": 9,
				"revpos": 1,
				"stub": true
			},
			"$main.css": {
				"content_type": "text/css",
				"digest": "md5-DUokkbXkY57LnAfJ6uvUTA==",
				"length": 111,
				"revpos": 1,
				"stub": true
			},
			"$template.0.html": {
				"content_type": "text/html",
				"digest": "md5-gyH1Ahu+xfkDHf8Y5lydSA==",
				"length": 348,
				"revpos": 1,
				"stub": true
			}
		},
		"_id": "theme-uQ3TFsQgm9Y29vlgC-lphauhK3M",
		"_rev": "1-33dac82cfe23d7773ea07c71f21cce38",
		"created": "2015-09-06T17:04:36.000000823Z",
		"files": [
			"$main.css"
		],
		"imported": "2016-09-06T20:22:40.438227326+02:00",
		"modelSequence": 1,
		"models": [
			{
				"fields": [
					{
						"fieldType": 3,
						"name": "Front"
					},
					{
						"fieldType": 3,
						"name": "Back"
					}
				],
				"files": [
					"!Basic-24b78.Card 1 answer.html",
					"!Basic-24b78.Card 1 question.html",
					"!Basic-24b78.Card 2 answer.html",
					"!Basic-24b78.Card 2 question.html",
					"$template.0.html"
				],
				"id": 0,
				"modelType": 0,
				"name": "Basic-24b78",
				"templates": [
					"Card 1",
					"Card 2"
				]
			}
		],
		"modified": "2016-08-02T13:15:15Z",
		"name": "Basic-24b78",
		"type": "theme"
	}`),
}

// Expected returns an object representing an expected document, replacing the
// rev with the one provided (for consistency)
func Expected(id, rev string) interface{} {
	doc, ok := expected[id]
	if !ok {
		panic(fmt.Sprintf("Expected doc '%s' not found", id))
	}
	var i interface{}
	if err := json.Unmarshal(doc, &i); err != nil {
		panic(fmt.Sprintf("Error unmarshaling expected doc '%s': %s", id, err))
	}
	i.(map[string]interface{})["_rev"] = rev
	return i
}
