package fb

import (
	"encoding/json"
	"testing"

	"github.com/flimzy/diff"
)

func TestNewTheme(t *testing.T) {
	type Test struct {
		name     string
		id       string
		expected interface{}
		err      string
	}
	tests := []Test{
		{
			name: "no id",
			err:  "id required",
		},
		{
			name: "valid",
			id:   "theme-foo",
			expected: func() *Theme {
				t := &Theme{
					ID:          "theme-foo",
					Created:     now(),
					Modified:    now(),
					Models:      make([]*Model, 0, 1),
					Attachments: NewFileCollection(),
				}
				t.Files = t.Attachments.NewView()
				return t
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := NewTheme(test.id)
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

func TestSetFile(t *testing.T) {
	att := NewFileCollection()
	view := att.NewView()
	_ = view.AddFile("foo.mp3", "audio/mpeg", []byte("foo"))
	theme, _ := NewTheme("theme-foo")
	theme.SetFile("foo.mp3", "audio/mpeg", []byte("foo"))
	expected := &Theme{
		ID:          "theme-foo",
		Created:     now(),
		Modified:    now(),
		Models:      []*Model{},
		Attachments: att,
		Files:       view,
	}
	if d := diff.Interface(expected, theme); d != nil {
		t.Error(d)
	}
}

func TestThemeMarshalJSON(t *testing.T) {
	type Test struct {
		name     string
		theme    *Theme
		expected string
		err      string
	}
	tests := []Test{
		{
			name:  "fails validation",
			theme: &Theme{},
			err:   "json: error calling MarshalJSON for type *fb.Theme: id required",
		},
		{
			name: "null fields",
			theme: func() *Theme {
				theme, _ := NewTheme("theme-abcd")
				theme.SetFile("file.txt", "text/plain", []byte("some text"))
				theme.Created = now()
				theme.Modified = now()
				return theme
			}(),
			expected: `{
				"_id":           "theme-abcd",
				"type":          "theme",
				"created":       "2017-01-01T00:00:00Z",
				"modified":      "2017-01-01T00:00:00Z",
				"modelSequence": 0,
				"files":         ["file.txt"],
				"_attachments":  {
					"file.txt": {
						"content_type": "text/plain",
						"data":         "c29tZSB0ZXh0"
					}
				}
			}`,
		},
		{
			name: "full fields",
			theme: func() *Theme {
				theme, _ := NewTheme("theme-abcd")
				theme.SetFile("file.txt", "text/plain", []byte("some text"))
				theme.Created = now()
				theme.Modified = now()
				theme.Imported = now()
				theme.Name = "Test Theme"
				theme.Description = "Theme for testing"
				theme.ModelSequence = 1
				m := &Model{
					Theme: theme,
					Type:  "foo",
					Files: theme.Attachments.NewView(),
				}
				theme.Models = []*Model{m}
				return theme
			}(),
			expected: `{
				"_id":           "theme-abcd",
				"type":          "theme",
				"name":          "Test Theme",
				"description":   "Theme for testing",
				"created":       "2017-01-01T00:00:00Z",
				"modified":      "2017-01-01T00:00:00Z",
				"imported":      "2017-01-01T00:00:00Z",
				"modelSequence": 1,
				"files":         ["file.txt"],
				"_attachments":  {
					"file.txt": {
						"content_type": "text/plain",
						"data":         "c29tZSB0ZXh0"
					}
				},
				"models": [
					{
						"fields": null,
						"files": [],
						"id": 0,
						"modelType": "foo",
						"templates": null
					}
				]
			}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := json.Marshal(test.theme)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.JSON([]byte(test.expected), result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestThemeUnmarshalJSON(t *testing.T) {
	type Test struct {
		name     string
		input    string
		expected *Theme
		err      string
	}
	tests := []Test{
		{
			name:  "invalid json",
			input: "xxx",
			err:   "failed to unmarshal Theme: invalid character 'x' looking for beginning of value",
		},
		{
			name:  "no attachments",
			input: `{"_id":"theme-120","created":"2017-01-01T01:01:01Z","modified":"2017-01-01T01:01:01Z"}`,
			err:   "invalid theme: no attachments",
		},
		{
			name: "with attachments",
			input: `{"_id":"theme-120", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z", "_attachments":{
			"foo.txt": {"content_type":"text/plain", "content": "text"}
			}}`,
			err: "invalid theme: no file list",
		},
		{
			name:  "mismatched file list",
			input: `{"_id":"theme-120", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z", "_attachments": {"foo.txt": {"content_type":"text/plain", "content": "text"}}, "files": ["foo.html"] }`,
			err:   "foo.html not found in collection",
		},
		{
			name:  "mismatched model file list",
			input: `{"_id":"theme-120", "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z", "_attachments": {"foo.txt": {"content_type":"text/plain", "content": "text"}}, "files":[], "models": [{"id":0, "files": ["foo.mp3"]}] }`,
			err:   "foo.mp3 not found in collection",
		},
		{
			name: "null fields",
			input: `{
				"_id":           "theme-abcd",
				"created":       "2017-01-01T00:00:00Z",
				"modified":      "2017-01-01T00:00:00Z",
				"modelSequence": 0,
				"files":         ["file.txt"],
				"_attachments":  {
					"file.txt": {
						"content_type": "text/plain",
						"data":         "c29tZSB0ZXh0"
					}
				}
			}`,
			expected: func() *Theme {
				theme, _ := NewTheme("theme-abcd")
				theme.SetFile("file.txt", "text/plain", []byte("some text"))
				theme.Created = now()
				theme.Modified = now()
				return theme
			}(),
		},
		{
			name: "full fields",
			input: `{
				"_id":           "theme-abcd",
				"name":          "Test Theme",
				"description":   "Theme for testing",
				"created":       "2017-01-01T00:00:00Z",
				"modified":      "2017-01-01T00:00:00Z",
				"imported":      "2017-01-01T00:00:00Z",
				"modelSequence": 1,
				"files":         ["file.txt"],
				"_attachments":  {
					"file.txt": {
						"content_type": "text/plain",
						"data":         "c29tZSB0ZXh0"
					}
				},
				"models": [
					{
						"fields": null,
						"files": [],
						"id": 0,
						"modelType": "foo",
						"templates": null
					}
				]
			}`,
			expected: func() *Theme {
				theme, _ := NewTheme("theme-abcd")
				theme.SetFile("file.txt", "text/plain", []byte("some text"))
				theme.Created = now()
				theme.Modified = now()
				theme.Imported = now()
				theme.Name = "Test Theme"
				theme.Description = "Theme for testing"
				theme.ModelSequence = 1
				m := &Model{
					Theme: theme,
					Type:  "foo",
					Files: theme.Attachments.NewView(),
				}
				theme.Models = []*Model{m}
				return theme
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := &Theme{}
			err := result.UnmarshalJSON([]byte(test.input))
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if d := diff.AsJSON(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestThemeNewModel(t *testing.T) {
	type Test struct {
		name      string
		theme     *Theme
		modelType string
		expected  *Model
		err       string
	}
	tests := []Test{
		{
			name:  "no type",
			theme: &Theme{},
			err:   "failed to create model: model type is required",
		},
		{
			name: "success",
			theme: func() *Theme {
				theme, _ := NewTheme("theme-foo")
				return theme
			}(),
			modelType: "chicken",
			expected: func() *Model {
				theme, _ := NewTheme("theme-foo")
				theme.ModelSequence = 1
				model := &Model{
					Type:      "chicken",
					Templates: []string{},
					Fields:    []*Field{},
					Files:     theme.Attachments.NewView(),
					Theme:     theme,
				}
				theme.Models = []*Model{model}
				return model
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.theme.NewModel(test.modelType)
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

func TestThemeSetRev(t *testing.T) {
	theme := &Theme{}
	rev := "1-xxx"
	theme.SetRev(rev)
	if theme.Rev != rev {
		t.Errorf("failed to set rev")
	}
}

func TestThemeID(t *testing.T) {
	expected := "theme-Zm9v"
	theme := &Theme{ID: expected}
	if id := theme.DocID(); id != expected {
		t.Errorf("unexpected id: %s", id)
	}
}

func TestThemeImportedTime(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		theme := &Theme{}
		ts := now()
		theme.Imported = ts
		if it := theme.ImportedTime(); it != ts {
			t.Errorf("Unexpected result: %s", it)
		}
	})
	t.Run("Unset", func(t *testing.T) {
		theme := &Theme{}
		if it := theme.ImportedTime(); !it.IsZero() {
			t.Errorf("unexpected result: %v", it)
		}
	})
}

func TestThemeModifiedTime(t *testing.T) {
	theme := &Theme{}
	ts := now()
	theme.Modified = ts
	if mt := theme.ModifiedTime(); mt != ts {
		t.Errorf("Unexpected result")
	}
}

func TestThemeMergeImport(t *testing.T) {
	type Test struct {
		name          string
		new           *Theme
		existing      *Theme
		expected      bool
		expectedTheme *Theme
		err           string
	}
	tests := []Test{
		{
			name:     "different ids",
			new:      &Theme{ID: "theme-abcd"},
			existing: &Theme{ID: "theme-b"},
			err:      "IDs don't match",
		},
		{
			name:     "created timestamps don't match",
			new:      &Theme{ID: "theme-abcd", Created: parseTime("2017-01-01T01:01:01Z"), Imported: parseTime("2017-01-15T00:00:00Z")},
			existing: &Theme{ID: "theme-abcd", Created: parseTime("2017-02-01T01:01:01Z"), Imported: parseTime("2017-01-20T00:00:00Z")},
			err:      "Created timestamps don't match",
		},
		{
			name:     "new not an import",
			new:      &Theme{ID: "theme-abcd", Created: parseTime("2017-01-01T01:01:01Z")},
			existing: &Theme{ID: "theme-abcd", Created: parseTime("2017-01-01T01:01:01Z"), Imported: parseTime("2017-01-15T00:00:00Z")},
			err:      "not an import",
		},
		{
			name:     "existing not an import",
			new:      &Theme{ID: "theme-abcd", Created: parseTime("2017-01-01T01:01:01Z"), Imported: parseTime("2017-01-15T00:00:00Z")},
			existing: &Theme{ID: "theme-abcd", Created: parseTime("2017-01-01T01:01:01Z")},
			err:      "not an import",
		},
		{
			name: "new is newer",
			new: &Theme{
				ID:       "theme-abcd",
				Name:     "foo",
				Created:  parseTime("2017-01-01T01:01:01Z"),
				Modified: parseTime("2017-02-01T01:01:01Z"),
				Imported: parseTime("2017-01-15T00:00:00Z"),
			},
			existing: &Theme{
				ID:       "theme-abcd",
				Name:     "bar",
				Created:  parseTime("2017-01-01T01:01:01Z"),
				Modified: parseTime("2017-01-01T01:01:01Z"),
				Imported: parseTime("2017-01-20T00:00:00Z"),
			},
			expected: true,
			expectedTheme: &Theme{
				ID:       "theme-abcd",
				Name:     "foo",
				Created:  parseTime("2017-01-01T01:01:01Z"),
				Modified: parseTime("2017-02-01T01:01:01Z"),
				Imported: parseTime("2017-01-15T00:00:00Z"),
			},
		},
		{
			name: "existing is newer",
			new: &Theme{
				ID:       "theme-abcd",
				Name:     "foo",
				Created:  parseTime("2017-01-01T01:01:01Z"),
				Modified: parseTime("2017-01-01T01:01:01Z"),
				Imported: parseTime("2017-01-15T00:00:00Z"),
			},
			existing: &Theme{
				ID:       "theme-abcd",
				Name:     "bar",
				Created:  parseTime("2017-01-01T01:01:01Z"),
				Modified: parseTime("2017-02-01T01:01:01Z"),
				Imported: parseTime("2017-01-20T00:00:00Z"),
			},
			expected: false,
			expectedTheme: &Theme{
				ID:       "theme-abcd",
				Name:     "bar",
				Created:  parseTime("2017-01-01T01:01:01Z"),
				Modified: parseTime("2017-02-01T01:01:01Z"),
				Imported: parseTime("2017-01-20T00:00:00Z"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.new.MergeImport(test.existing)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if test.expected != result {
				t.Errorf("Unexpected result: %t", result)
			}
			if d := diff.Interface(test.expectedTheme, test.new); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestThemeValidate(t *testing.T) {
	att := NewFileCollection()
	view := att.NewView()
	tests := []validationTest{
		{
			name: "no ID",
			v:    &Theme{},
			err:  "id required",
		},
		{
			name: "invalid doctype",
			v:    &Theme{ID: "chicken-a"},
			err:  "unsupported DocID type 'chicken'",
		},
		{
			name: "wrong doctype",
			v:    &Theme{ID: "deck-abcd"},
			err:  "incorrect doc type",
		},
		{
			name: "no created time",
			v:    &Theme{ID: "theme-abcd"},
			err:  "created time required",
		},
		{
			name: "no modified time",
			v:    &Theme{ID: "theme-abcd", Created: now()},
			err:  "modified time required",
		},
		{
			name: "nil attachments collection",
			v:    &Theme{ID: "theme-abcd", Created: now(), Modified: now()},
			err:  "attachments collection must not be nil",
		},
		{
			name: "nil file list",
			v:    &Theme{ID: "theme-abcd", Created: now(), Modified: now(), Attachments: NewFileCollection()},
			err:  "file list must not be nil",
		},
		{
			name: "attachments and files don't match",
			v:    &Theme{ID: "theme-abcd", Created: now(), Modified: now(), Attachments: NewFileCollection(), Files: NewFileCollection().NewView()},
			err:  "file list must be a member of attachments collection",
		},
		{
			name: "invalid model sequence",
			v:    &Theme{ID: "theme-abcd", Created: now(), Modified: now(), Attachments: att, Files: view, ModelSequence: 0, Models: []*Model{{ID: 0}}},
			err:  "modelSequence must be larger than existing model IDs",
		},
		{
			name: "invalid model file list",
			v:    &Theme{ID: "theme-abcd", Created: now(), Modified: now(), Attachments: att, Files: view, ModelSequence: 1, Models: []*Model{{ID: 0, Files: NewFileCollection().NewView()}}},
			err:  "model 0 file list must be a member of attachments collection",
		},
		{
			name: "invalid model",
			v:    &Theme{ID: "theme-abcd", Created: now(), Modified: now(), Attachments: att, Files: view, ModelSequence: 1, Models: []*Model{{ID: 0, Files: view}}},
			err:  "invalid model: theme is required",
		},
		{
			name: "valid",
			v:    &Theme{ID: "theme-abcd", Created: now(), Modified: now(), Attachments: att, Files: view},
		},
	}
	testValidation(t, tests)
}
