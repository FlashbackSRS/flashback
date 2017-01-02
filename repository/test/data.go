package test

import (
	"encoding/json"
	"fmt"

	"github.com/pborman/uuid"

	"github.com/FlashbackSRS/flashback-model"
)

var testUser *fb.User

func init() {
	u, err := fb.NewUser(uuid.Parse("9d11d024-a100-4045-a5b7-9f1ccf96cc9f"), "mrsmith")
	if err != nil {
		panic(fmt.Sprintf("Error creating user: %s\n", err))
	}
	testUser = u
}

var expected = map[string][]byte{
	"theme-ELr8cEJJOvJU4lYz-VTXhH8wLTo": []byte(`{
		"_attachments": {
			"!Basic-24b78.Card 1 answer.html": {
				"content_type": "text/html",
				"digest": "md5-ACvi0DsrBFAgRdOV9DdRmg==",
				"length": 72,
				"revpos": 1,
				"stub": true
			},
			"!Basic-24b78.Card 1 question.html": {
				"content_type": "text/html",
				"digest": "md5-kzMtJPyK2E4mIieI4jEzJQ==",
				"length": 20,
				"revpos": 1,
				"stub": true
			},
			"!Basic-24b78.Card 2 answer.html": {
				"content_type": "text/html",
				"digest": "md5-t4DtgpoCeNyU6yT9O7bxWw==",
				"length": 39,
				"revpos": 1,
				"stub": true
			},
			"!Basic-24b78.Card 2 question.html": {
				"content_type": "text/html",
				"digest": "md5-/s7RZUPkkLFG6JwFsSWiag==",
				"length": 19,
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
		"_id": "theme-ELr8cEJJOvJU4lYz-VTXhH8wLTo",
		"_rev": "1-33dac82cfe23d7773ea07c71f21cce38",
		"created": "2015-09-06T17:04:36.000000823Z",
		"files": [
			"$main.css"
		],
		"imported": "2017-01-02T17:16:56.764985035+01:00",
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
				"modelType": "anki-basic",
				"name": "Basic-24b78",
				"templates": [
					"Card 1",
					"Card 2"
				]
			}
		],
		"modified": "2016-09-11T19:01:39Z",
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
