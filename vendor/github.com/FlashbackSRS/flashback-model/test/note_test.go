package test

import (
	"encoding/json"
	"testing"

	"github.com/flimzy/testify/require"

	"github.com/FlashbackSRS/flashback-model"
)

var frozenNote = []byte(`
{
    "_id": "note-VGVzdCBOb3Rl",
	"type": "note",
    "created": "2016-07-31T15:08:24.730156517Z",
    "modified": "2016-07-31T15:08:24.730156517Z",
    "imported": "2016-08-02T15:08:24.730156517Z",
    "theme": "theme-VGVzdCBUaGVtZQ",
    "model": 1,
    "fieldValues": [
        {
            "text": "cat"
        },
        {
            "files": [
                "_weirdname.txt",
                "foo.mp3",
                "영상.jpg"
            ]
        }
    ],
    "_attachments": {
        "영상.jpg": {
            "content_type": "audio/mpeg",
            "data": "YSBLb3JlYW4gZmlsZW5hbWU="
        },
        "^_weirdname.txt": {
            "content_type": "audio/mpeg",
            "data": "YSBmaWxlIHdpdGggYSBzdHJhbmdlIG5hbWU="
        },
        "foo.mp3": {
            "content_type": "audio/mpeg",
            "data": "bm90IGEgcmVhbCBNUDM="
        }
    }
}
`)

var frozenExistingNote = []byte(`
{
    "_id": "note-VGVzdCBOb3Rl",
    "_rev": "1-6e1b6fb5352429cf3013eab5d692aac8",
    "created": "2016-07-31T15:08:24.730156517Z",
    "modified": "2016-07-15T15:07:24.730156517Z",
    "imported": "2016-08-01T15:08:24.730156517Z",
    "theme": "theme-VGVzdCBUaGVtZQ",
    "model": 1,
    "fieldValues": [
        {
            "text": "Cat"
        },
        {
            "files": [
                "foo.mp3"
            ]
        }
    ],
    "_attachments": {
        "foo.mp3": {
            "content_type": "audio/mpeg",
            "data": "bm90IGEgcmVhbCBNUDM="
        }
    }
}
`)

var frozenMergedNote = []byte(`
{
    "_id": "note-VGVzdCBOb3Rl",
	"type": "note",
    "_rev": "1-6e1b6fb5352429cf3013eab5d692aac8",
    "created": "2016-07-31T15:08:24.730156517Z",
    "modified": "2016-07-31T15:08:24.730156517Z",
    "imported": "2016-08-02T15:08:24.730156517Z",
    "theme": "theme-VGVzdCBUaGVtZQ",
    "model": 1,
    "fieldValues": [
        {
            "text": "cat"
        },
        {
            "files": [
                "_weirdname.txt",
                "foo.mp3",
                "영상.jpg"
            ]
        }
    ],
    "_attachments": {
        "^_weirdname.txt": {
            "content_type": "audio/mpeg",
            "data": "YSBmaWxlIHdpdGggYSBzdHJhbmdlIG5hbWU="
        },
        "영상.jpg": {
            "content_type": "audio/mpeg",
            "data": "YSBLb3JlYW4gZmlsZW5hbWU="
        },
        "foo.mp3": {
            "content_type": "audio/mpeg",
            "data": "bm90IGEgcmVhbCBNUDM="
        }
    }
}
`)

func TestNoteMergeImport(t *testing.T) {
	require := require.New(t)
	th := &fb.Theme{}
	json.Unmarshal(frozenTheme, th)
	m := th.Models[1]
	n := &fb.Note{}
	err := json.Unmarshal(frozenNote, n)
	require.Nil(err, "Error thawing Note: %s", err)

	n.SetModel(m)
	e := &fb.Note{}
	err = json.Unmarshal(frozenExistingNote, e)
	require.Nil(err, "Error thawing ExistingNote: %s", err)

	e.SetModel(m)
	changed, err := n.MergeImport(e)
	require.Nil(err, "Error merging Note: %s", err)
	require.True(changed, "No change!")
	require.MarshalsToJSON(frozenMergedNote, n, "Merged Note")
}
