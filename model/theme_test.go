package model

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"strings"
	"testing"

	fb "github.com/FlashbackSRS/flashback-model"
	"github.com/FlashbackSRS/flashback/controllers"
	// _ "github.com/FlashbackSRS/flashback/controllers/anki"
	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
)

type basicCM struct {
	controllers.ModelController
}

func (cm *basicCM) Type() string         { return "basic" }
func (cm *basicCM) IframeScript() []byte { return []byte("alert('Hi!');") }

var _ controllers.ModelController = &basicCM{}

func init() {
	controllers.RegisterModelController(&basicCM{})
}

var theme1 = func() *fb.Theme {
	body := []byte(`
    {
        "type": "theme",
        "_id": "theme-VGVzdCBUaGVtZQ",
        "created": "2016-07-31T15:08:24.730156517Z",
        "modified": "2016-07-31T15:08:24.730156517Z",
        "imported": "2016-08-02T15:08:24.730156517Z",
        "name": "Test Theme",
        "description": "Theme for testing",
        "models": [
            {
                "id": 0,
                "modelType": "basic",
                "name": "Model A",
                "templates": ["Card 1"],
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
                    "$template.0.html",
                    "m1.html"
                ]
            },
            {
                "id": 1,
                "modelType": "basic",
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
            },
            {
                "id": 2,
                "modelType": "unknownType",
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
                    "$template.2.html",
                    "m1.txt"
                ]
            },
            {
                "id": 3,
                "modelType": "basic",
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
                    "$template.3.html",
                    "m1.txt"
                ]
            }
        ],
        "_attachments": {
            "$template.0.html": {
                "content_type": "text/html",
                "data": "Qm9yaW5nIHRlbXBsYXRlCg=="
            },
            "$template.2.html": {
                "content_type": "text/html",
                "data": "Qm9yaW5nIHRlbXBsYXRlCg=="
            },
            "$template.3.html": {
                "content_type": "text/html",
                "data": "e3sK"
            },
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
        "modelSequence": 4
    }
    `)
	theme := &fb.Theme{}
	if err := json.Unmarshal(body, &theme); err != nil {
		panic(err)
	}
	return theme
}()

var theme2 = func() *fb.Theme {
	body := []byte(`
    {
        "type": "theme",
        "_id": "theme-VGVzdCBUaGVtZQ",
        "created": "2016-07-31T15:08:24.730156517Z",
        "modified": "2016-07-31T15:08:24.730156517Z",
        "imported": "2016-08-02T15:08:24.730156517Z",
        "name": "Test Theme",
        "description": "Theme for testing",
        "models": [
            {
                "id": 0,
                "modelType": "basic",
                "name": "Model A",
                "templates": ["Card 1"],
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
                    "$template.0.html"
                ]
            }
        ],
        "_attachments": {
            "$template.0.html": {
                "content_type": "text/html",
                "data": "Qm9yaW5nIHRlbXBsYXRlCg=="
            },
            "$main.css": {
                "content_type": "text/css",
                "data": "e3sK"
            }
        },
        "files": [
            "$main.css"
        ],
        "modelSequence": 2
    }
    `)
	theme := &fb.Theme{}
	if err := json.Unmarshal(body, &theme); err != nil {
		panic(err)
	}
	return theme
}()

func TestModelTemplate(t *testing.T) {
	type mtTest struct {
		name     string
		theme    *fb.Theme
		modelID  int
		expected string
		err      string
	}
	tests := []mtTest{
		{
			name:    "main template missing",
			theme:   theme1,
			modelID: 1,
			err:     "main template '$template.1.html' not found in model",
		},
		{
			name:    "unknown model type",
			theme:   theme1,
			modelID: 2,
			err:     "ModelController for 'unknownType' not found",
		},
		{
			name:    "invalid template",
			theme:   theme1,
			modelID: 3,
			err:     "Error parsing template file `template.html`: template: template:2: unexpected unclosed action in command",
		},
		{
			name:    "invalid css",
			theme:   theme2,
			modelID: 0,
			err:     "failed to parse $main.css: template: template:2: unexpected unclosed action in command",
		},
		{
			name:    "single anki template",
			theme:   theme1,
			modelID: 0,
			expected: `<!DOCTYPE html>
<html>
<head>
	<title>FB Card</title>
	<base href="">
	<meta charset="UTF-8">
	<link rel="stylesheet" type="text/css" href="css/cardframe.css">
<script type="text/javascript">
'use strict';
var FB = {
	face: "",
	card: "",
	note: ""
};
</script>
<script type="text/javascript" src="js/cardframe.js"></script>
<script type="text/javascript"></script>
<style> </style>
</head>
<body>Boring template
</body>
</html>
`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			model := &fbModel{
				Model: test.theme.Models[test.modelID],
			}
			result, err := model.Template(context.Background())
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			buf := &bytes.Buffer{}
			if e := result.Execute(buf, nil); e != nil {
				panic(e)
			}
			if d := diff.Text(test.expected, buf.String()); d != nil {
				t.Error(d)
			}
		})
	}
}

type funcMapperCM struct {
	controllers.ModelController
}

func (cm *funcMapperCM) Type() string { return "funcmapper" }
func (cm *funcMapperCM) FuncMap(card *fb.Card, face int) template.FuncMap {
	return map[string]interface{}{
		"foo": nilFunc,
	}
}

var _ controllers.ModelController = &funcMapperCM{}
var _ controllers.FuncMapper = &funcMapperCM{}

func init() {
	controllers.RegisterModelController(&funcMapperCM{})
}

var nilFunc = func() {}

func TestFuncMap(t *testing.T) {
	tests := []struct {
		name      string
		modelType string
		card      *fbCard
		expected  template.FuncMap
		err       string
	}{
		{
			name:      "unregistered",
			modelType: "unregistered",
			err:       "ModelController for 'unregistered' not found",
		},
		{
			name:      "non funcMapper",
			modelType: "basic",
		},
		{
			name:      "funcMapper",
			modelType: "funcmapper",
			expected: template.FuncMap{
				"foo": nilFunc,
			},
		},
		{
			name:      "non nil card",
			modelType: "funcmapper",
			card:      &fbCard{Card: &fb.Card{}},
			expected: template.FuncMap{
				"foo": nilFunc,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := &fbModel{Model: &fb.Model{Type: test.modelType}}
			result, err := m.FuncMap(test.card, 0)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestIframeScript(t *testing.T) {
	t.Run("unregistered", func(t *testing.T) {
		m := &fbModel{Model: &fb.Model{Type: "unregistered"}}
		_, err := m.IframeScript()
		checkErr(t, "ModelController for 'unregistered' not found", err)
	})
	t.Run("registered", func(t *testing.T) {
		m := &fbModel{Model: &fb.Model{Type: "basic"}}
		s, err := m.IframeScript()
		checkErr(t, nil, err)
		expected := "alert('Hi!');"
		if string(s) != expected {
			t.Errorf("Unexpected result:\n%s\n", string(s))
		}
	})
}

func TestGetAttachment(t *testing.T) {
	tests := []struct {
		name     string
		model    *fbModel
		filename string
		expected *fb.Attachment
		err      string
	}{
		{
			name: "content already set",
			model: func() *fbModel {
				theme, err := fb.NewTheme("theme-aaa")
				if err != nil {
					t.Fatal(err)
				}
				model, err := fb.NewModel(theme, "basic")
				if err != nil {
					t.Fatal(err)
				}
				if err := model.Files.AddFile("foo.txt", "text/plain", []byte("some text")); err != nil {
					t.Fatal(err)
				}
				return &fbModel{
					Model: model,
				}
			}(),
			filename: "foo.txt",
			expected: &fb.Attachment{
				ContentType: "text/plain",
				Content:     []byte("some text"),
			},
		},
		{
			name:     "not found",
			model:    &fbModel{Model: realTheme.Models[0]},
			filename: "foo.txt",
			err:      "attachment 'foo.txt' not found",
		},
		{
			name: "db error",
			model: &fbModel{
				Model: realTheme.Models[0],
				db: &mockAttachmentGetter{
					err: errors.New("db error"),
				},
			},
			filename: "!Basic-24b78.Card 1 answer.html",
			err:      "db error",
		},
		{
			name: "db fetch",
			model: &fbModel{
				Model: realTheme.Models[0],
				db: &mockAttachmentGetter{
					attachments: map[string]*kivik.Attachment{
						"!Basic-24b78.Card 1 answer.html": {
							Filename:    "!Basic-24b78.Card 1 answer.html",
							ContentType: "text/html",
							ReadCloser:  ioutil.NopCloser(strings.NewReader("some html")),
						},
					},
				},
			},
			filename: "!Basic-24b78.Card 1 answer.html",
			expected: &fb.Attachment{
				ContentType: "text/html",
				Content:     []byte("some html"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.model.getAttachment(context.Background(), test.filename)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
