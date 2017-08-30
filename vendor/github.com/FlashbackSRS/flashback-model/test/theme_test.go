package test

import (
	"encoding/json"
	"testing"

	"github.com/flimzy/testify/require"

	"github.com/FlashbackSRS/flashback-model"
)

var frozenTheme = []byte(`
{
    "_id": "theme-VGVzdCBUaGVtZQ",
	"type": "theme",
    "created": "2016-07-31T15:08:24.730156517Z",
    "modified": "2016-07-31T15:08:24.730156517Z",
    "imported": "2016-08-02T15:08:24.730156517Z",
    "name": "Test Theme",
    "description": "Theme for testing",
    "models": [
        {
            "id": 0,
            "modelType": "anki-basic",
            "name": "Model A",
            "templates": [],
            "fields": [
                {
                    "fieldType": 0,
                    "name": "Word"
                },
                {
                    "fieldType": 0,
                    "name": "Definition"
                }
            ],
            "files": [
                "m1.html"
            ]
        },
        {
            "id": 1,
            "modelType": "anki-cloze",
            "name": "Model 2",
            "templates": [],
            "fields": [
                {
                    "fieldType": 0,
                    "name": "Word"
                },
                {
                    "fieldType": 2,
                    "name": "Audio"
                }
            ],
            "files": [
                "m1.txt"
            ]
        }
    ],
    "_attachments": {
        "m1.html": {
            "content_type": "text/html",
            "data": "PGh0bWw+PC9odG1sPg=="
        },
        "m1.txt": {
            "content_type": "text/plain",
            "data": "VGVzdCB0ZXh0IGZpbGU="
        },
        "$main.css": {
            "content_type": "text/css",
            "data": "LyogYW4gZW1wdHkgQ1NTIGZpbGUgKi8="
        }
    },
    "files": [
        "$main.css"
    ],
    "modelSequence": 2
}
`)

func TestCreateTheme(t *testing.T) {
	require := require.New(t)
	th, err := fb.NewTheme("theme-VGVzdCBUaGVtZQ")
	require.Nil(err, "Error creating theme: %s", err)

	th.Name = "Test Theme"
	th.Description = "Theme for testing"
	th.Created = now
	th.Modified = now
	th.Imported = now.AddDate(0, 0, 2)
	th.SetFile("$main.css", "text/css", []byte("/* an empty CSS file */"))
	m1, _ := th.NewModel("anki-basic")
	m2, _ := th.NewModel("anki-cloze")
	m1.AddField(fb.TextField, "Word")
	m1.AddField(fb.TextField, "Definition")
	m2.AddField(fb.TextField, "Word")
	m2.AddField(fb.AudioField, "Audio")
	m1.Name = "Model A"
	m2.Name = "Model 2"
	m1.AddFile("m1.html", "text/html", []byte("<html></html>"))
	m2.AddFile("m1.txt", "text/plain", []byte("Test text file"))
	require.MarshalsToJSON(frozenTheme, th, "Create Theme")

	th2 := &fb.Theme{}
	err = json.Unmarshal(frozenTheme, th2)
	require.Nil(err, "Error thawing theme: %s", err)
	require.MarshalsToJSON(frozenTheme, th2, "Thawed Theme")

	require.DeepEqual(th, th2, "Thawed vs. Created Themes")
}

var frozenExistingTheme = []byte(`
{
    "_id": "theme-VGVzdCBUaGVtZQ",
    "_rev": "1-6e1b6fb5352429cf3013eab5d692aac8",
    "created": "2016-07-31T15:08:24.730156517Z",
    "modified": "2016-07-15T15:07:24.730156517Z",
    "imported": "2016-08-01T15:08:24.730156517Z",
    "name": "Test Theme",
    "description": "Theme for testing",
    "models": [
        {
            "id": 0,
            "modelType": "anki-basic",
            "name": "Model A",
            "templates": [],
            "fields": [
                {
                    "fieldType": 0,
                    "name": "Word"
                },
                {
                    "fieldType": 0,
                    "name": "Definition"
                }
            ],
            "files": [
                "m1.html"
            ]
        },
        {
            "id": 1,
            "modelType": "anki-cloze",
            "name": "Model 2",
            "templates": [],
            "fields": [
                {
                    "fieldType": 0,
                    "name": "Word"
                },
                {
                    "fieldType": 2,
                    "name": "Audio"
                }
            ],
            "files": [
                "m1.txt"
            ]
        }
    ],
    "_attachments": {
        "m1.html": {
            "content_type": "text/html",
            "data": "PGh0bWw+PC9odG1sPg=="
        },
        "m1.txt": {
            "content_type": "text/plain",
            "data": "VGVzdCB0ZXh0IGZpbGU="
        },
        "$main.css": {
            "content_type": "text/css",
            "data": "LyogYW4gZW1wdHkgQ1NTIGZpbGUgKi8="
        }
    },
    "files": [
        "$main.css"
    ],
    "modelSequence": 2
}
`)

var frozenMergedTheme = []byte(`
{
    "_id": "theme-VGVzdCBUaGVtZQ",
	"type": "theme",
    "_rev": "1-6e1b6fb5352429cf3013eab5d692aac8",
    "created": "2016-07-31T15:08:24.730156517Z",
    "modified": "2016-07-31T15:08:24.730156517Z",
    "imported": "2016-08-02T15:08:24.730156517Z",
    "name": "Test Theme",
    "description": "Theme for testing",
    "models": [
        {
            "id": 0,
            "modelType": "anki-basic",
            "name": "Model A",
            "templates": [],
            "fields": [
                {
                    "fieldType": 0,
                    "name": "Word"
                },
                {
                    "fieldType": 0,
                    "name": "Definition"
                }
            ],
            "files": [
                "m1.html"
            ]
        },
        {
            "id": 1,
            "modelType": "anki-cloze",
            "name": "Model 2",
            "templates": [],
            "fields": [
                {
                    "fieldType": 0,
                    "name": "Word"
                },
                {
                    "fieldType": 2,
                    "name": "Audio"
                }
            ],
            "files": [
                "m1.txt"
            ]
        }
    ],
    "_attachments": {
        "m1.html": {
            "content_type": "text/html",
            "data": "PGh0bWw+PC9odG1sPg=="
        },
        "m1.txt": {
            "content_type": "text/plain",
            "data": "VGVzdCB0ZXh0IGZpbGU="
        },
        "$main.css": {
            "content_type": "text/css",
            "data": "LyogYW4gZW1wdHkgQ1NTIGZpbGUgKi8="
        }
    },
    "files": [
        "$main.css"
    ],
    "modelSequence": 2
}
`)

func TestThemeMergeImport(t *testing.T) {
	require := require.New(t)
	th := &fb.Theme{}
	err := json.Unmarshal(frozenTheme, th)
	require.Nil(err, "Error thawing Theme: %s", err)

	e := &fb.Theme{}
	err = json.Unmarshal(frozenExistingTheme, e)
	require.Nil(err, "Error thawing ExistingTheme: %s", err)

	changed, err := th.MergeImport(e)
	require.Nil(err, "Error merging Theme: %s", err)
	require.True(changed, "No change merging Theme")

	require.MarshalsToJSON(frozenMergedTheme, th, "Merged Theme")
}
