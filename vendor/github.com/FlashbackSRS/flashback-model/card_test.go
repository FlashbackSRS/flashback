package fb

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/flimzy/diff"
)

func TestParseID(t *testing.T) {
	type pidTest struct {
		name              string
		input             string
		bundle, note, err string
		template          uint32
	}
	tests := []pidTest{
		{
			name: "empty input",
			err:  "invalid ID type",
		},
		{
			name:  "invalid ID",
			input: "card-the quick brown fox",
			err:   "invalid ID format",
		},
		{
			name:  "invalid template",
			input: "card-krsxg5baij2w4zdmmu.mViuXQThMLoh1G1Nlc4d_E8kR8o.boo",
			err:   `invalid TemplateID: strconv.Atoi: parsing "boo": invalid syntax`,
		},
		{
			name:  "wrong id type",
			input: "foo-card-krsxg5baij2w4zdmmu.mViuXQThMLoh1G1Nlc4d_E8kR8o.0",
			err:   "invalid ID type",
		},
		{
			name:     "valid",
			input:    "card-krsxg5baij2w4zdmmu.mViuXQThMLoh1G1Nlc4d_E8kR8o.0",
			bundle:   "krsxg5baij2w4zdmmu",
			note:     "mViuXQThMLoh1G1Nlc4d_E8kR8o",
			template: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := &Card{ID: test.input}
			bundle, note, template, err := c.parseID()
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if bundle != test.bundle || note != test.note || template != test.template {
				t.Errorf("Unexpected result: %s %s %d", bundle, note, template)
			}
		})
	}
}

func TestNewCard(t *testing.T) {
	type ncTest struct {
		name     string
		theme    string
		model    uint32
		id       string
		expected *Card
		err      string
	}
	tests := []ncTest{
		{
			name: "invalid id",
			id:   "chicken man",
			err:  "validation failure: invalid ID type",
		},
		{
			name:  "valid",
			theme: "theme-foo",
			id:    "card-krsxg5baij2w4zdmmu.mViuXQThMLoh1G1Nlc4d_E8kR8o.1",
			expected: &Card{
				ID:       "card-krsxg5baij2w4zdmmu.mViuXQThMLoh1G1Nlc4d_E8kR8o.1",
				ModelID:  "theme-foo/0",
				Created:  parseTime("2017-01-01T00:00:00Z"),
				Modified: parseTime("2017-01-01T00:00:00Z"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := NewCard(test.theme, test.model, test.id)
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

func TestCardMarshalJSON(t *testing.T) {
	t.Run("invalid", func(t *testing.T) {
		card := &Card{}
		_, err := json.Marshal(card)
		checkErr(t, "json: error calling MarshalJSON for type *fb.Card: validation error: id required", err)
	})
	t.Run("null fields", func(t *testing.T) {
		card := &Card{
			ID:       "card-foo.bar.1",
			ModelID:  "theme-baz/2",
			Created:  parseTime("2017-01-01T01:01:01Z"),
			Modified: parseTime("2017-01-01T01:01:01Z"),
		}
		expected := []byte(`{
			"type":     "card",
			"_id":      "card-foo.bar.1",
			"created":  "2017-01-01T01:01:01Z",
			"model":    "theme-baz/2",
			"modified": "2017-01-01T01:01:01Z"
		}`)
		result, err := json.Marshal(card)
		checkErr(t, nil, err)
		if d := diff.JSON(expected, result); d != nil {
			t.Error(d)
		}
	})
	t.Run("full fields", func(t *testing.T) {
		card := &Card{
			ID:          "card-foo.bar.1",
			ModelID:     "theme-baz/2",
			Created:     parseTime("2017-01-01T01:01:01Z"),
			Modified:    parseTime("2017-01-01T01:01:01Z"),
			Imported:    parseTime("2017-01-01T01:01:01Z"),
			BuriedUntil: Due(parseTime("2017-03-01T00:00:00Z")),
			Due:         Due(parseTime("2018-01-01T00:00:00Z")),
			LastReview:  parseTime("2016-12-30T12:00:00Z"),
			Suspended:   true,
		}
		expected := []byte(`{
			"type":        "card",
			"_id":         "card-foo.bar.1",
			"model":       "theme-baz/2",
			"created":     "2017-01-01T01:01:01Z",
			"modified":    "2017-01-01T01:01:01Z",
			"imported":    "2017-01-01T01:01:01Z",
			"lastReview":  "2016-12-30T12:00:00Z",
			"buriedUntil": "2017-03-01",
			"due":         "2018-01-01",
			"suspended":   true
		}`)
		result, err := json.Marshal(card)
		checkErr(t, nil, err)
		if d := diff.JSON(expected, result); d != nil {
			t.Error(d)
		}
	})
}

func TestUnmarshalJSON(t *testing.T) {
	type ujTest struct {
		name     string
		input    string
		expected *Card
		err      string
	}
	tests := []ujTest{
		{
			name: "no input",
			err:  "unexpected end of JSON input",
		},
		{
			name:  "validation failure",
			input: `{"_id":"oink"}`,
			err:   "validation error: invalid ID type",
		},
		{
			name:  "valid",
			input: `{"_id":"card-krsxg5baij2w4zdmmu.mViuXQThMLoh1G1Nlc4d_E8kR8o.1", "model": "theme-foo/2", "suspended":true, "created":"2017-01-01T01:01:01Z", "modified":"2017-01-01T01:01:01Z"}`,
			expected: &Card{
				ID:        "card-krsxg5baij2w4zdmmu.mViuXQThMLoh1G1Nlc4d_E8kR8o.1",
				ModelID:   "theme-foo/2",
				Created:   parseTime("2017-01-01T01:01:01Z"),
				Modified:  parseTime("2017-01-01T01:01:01Z"),
				Suspended: true,
			},
		},
		{
			name: "test frozen card",
			input: `
		{
			"_id": "card-krsxg5baij2w4zdmmu.mViuXQThMLoh1G1Nlc4d_E8kR8o.0",
			"created": "2016-07-31T15:08:24.730156517Z",
			"modified": "2016-07-31T15:08:24.730156517Z",
			"imported": "2016-08-02T15:08:24.730156517Z",
			"model": "theme-VGVzdCBUaGVtZQ/0",
			"due": "2017-01-01",
			"interval": 50
		}`,
			expected: &Card{
				ID:       "card-krsxg5baij2w4zdmmu.mViuXQThMLoh1G1Nlc4d_E8kR8o.0",
				ModelID:  "theme-VGVzdCBUaGVtZQ/0",
				Created:  parseTime("2016-07-31T15:08:24.730156517Z"),
				Modified: parseTime("2016-07-31T15:08:24.730156517Z"),
				Imported: parseTime("2016-08-02T15:08:24.730156517Z"),
				Interval: parseInterval("50d"),
				Due:      parseDue("2017-01-01"),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := &Card{}
			err := result.UnmarshalJSON([]byte(test.input))
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

func TestIdentity(t *testing.T) {
	card := &Card{ID: "card-bundle.note.2"}
	expected := "bundle.note.2"
	result := card.Identity()
	if result != expected {
		t.Errorf("Unexpected result: %s", result)
	}
}

func TestSetRev(t *testing.T) {
	card := &Card{}
	rev := "1-xxx"
	card.SetRev(rev)
	if card.Rev != rev {
		t.Errorf("Unexpected rev: %s", card.Rev)
	}
}

func TestDocID(t *testing.T) {
	card := &Card{ID: "card-bundle.note.2"}
	expected := "card-bundle.note.2"
	result := card.DocID()
	if result != expected {
		t.Errorf("Unexpected result: %s", result)
	}
}

func TestImportedTime(t *testing.T) {
	ts := time.Now()
	card := &Card{Imported: ts}
	if it := card.ImportedTime(); !it.Equal(ts) {
		t.Errorf("Unexpected result: %v", it)
	}
}

func TestModifiedTime(t *testing.T) {
	ts := time.Now()
	card := &Card{Modified: ts}
	if mt := card.ModifiedTime(); !mt.Equal(ts) {
		t.Errorf("Unexpected result: %v", mt)
	}
}

func TestMergeImport(t *testing.T) {
	type miTest struct {
		name         string
		card         *Card
		i            interface{}
		expected     bool
		expectedCard *Card
		err          string
	}
	tests := []miTest{
		{
			name: "no input",
			card: &Card{},
			i:    nil,
			err:  "i is <nil>, not *fb.Card",
		},
		{
			name: "mismatched identities",
			card: &Card{ID: "card-foo.bar.1"},
			i:    &Card{ID: "card-foo.bar.2"},
			err:  "IDs don't match",
		},
		{
			name: "different timestamps",
			card: &Card{ID: "card-foo.bar.1", Created: parseTime("2017-01-01T01:01:01Z"), Imported: parseTime("2017-01-15T00:00:00Z")},
			i:    &Card{ID: "card-foo.bar.1", Created: parseTime("2017-02-01T01:01:01Z"), Imported: parseTime("2017-01-20T00:00:00Z")},
			err:  "Created timestamps don't match",
		},
		{
			name: "new not an import",
			card: &Card{ID: "card-foo.bar.1", Created: parseTime("2017-01-01T01:01:01Z")},
			i:    &Card{ID: "card-foo.bar.1", Created: parseTime("2017-02-01T01:01:01Z"), Imported: parseTime("2017-01-20T00:00:00Z")},
			err:  "not an import",
		},
		{
			name: "existing not an import",
			card: &Card{ID: "card-foo.bar.1", Created: parseTime("2017-01-01T01:01:01Z"), Imported: parseTime("2017-01-20T00:00:00Z")},
			i:    &Card{ID: "card-foo.bar.1", Created: parseTime("2017-02-01T01:01:01Z")},
			err:  "not an import",
		},
		{
			name: "existing is newer",
			card: &Card{ID: "card-foo.bar.1",
				Created:  parseTime("2017-01-01T01:01:01Z"),
				Modified: parseTime("2017-01-01T01:01:01Z"),
				Imported: parseTime("2017-01-15T00:00:00Z")},
			i: &Card{ID: "card-foo.bar.1",
				Created:  parseTime("2017-01-01T01:01:01Z"),
				Modified: parseTime("2017-01-02T01:01:01Z"),
				Imported: parseTime("2017-01-20T00:00:00Z")},
			expectedCard: &Card{ID: "card-foo.bar.1",
				Created:  parseTime("2017-01-01T01:01:01Z"),
				Modified: parseTime("2017-01-02T01:01:01Z"),
				Imported: parseTime("2017-01-20T00:00:00Z")},
		},
		{
			name: "new is newer",
			card: &Card{ID: "card-foo.bar.1",
				Created:  parseTime("2017-01-01T01:01:01Z"),
				Modified: parseTime("2017-01-02T01:01:01Z"),
				Imported: parseTime("2017-01-15T00:00:00Z")},
			i: &Card{ID: "card-foo.bar.1",
				Created:  parseTime("2017-01-01T01:01:01Z"),
				Modified: parseTime("2017-01-01T01:01:01Z"),
				Imported: parseTime("2017-01-20T00:00:00Z")},
			expected: true,
			expectedCard: &Card{ID: "card-foo.bar.1",
				Created:  parseTime("2017-01-01T01:01:01Z"),
				Modified: parseTime("2017-01-02T01:01:01Z"),
				Imported: parseTime("2017-01-15T00:00:00Z")},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.card.MergeImport(test.i)
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if result != test.expected {
				t.Errorf("Unexpected result: %t", result)
			}
			if d := diff.Interface(test.expectedCard, test.card); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestCardBundleID(t *testing.T) {
	card := &Card{ID: "card-foo.bar.1"}
	expected := "bundle-foo"
	if id := card.BundleID(); id != expected {
		t.Errorf("Unexpected result: %s", id)
	}
}

func TestTemplateID(t *testing.T) {
	expected := uint32(3)
	card := &Card{ID: "card-foo.bar.3"}
	if id := card.TemplateID(); id != expected {
		t.Errorf("Unexpected result: %d", id)
	}
}

// func TestModelID(t *testing.T) {
// 	expected := 4
// 	card := &Card{ModelID: "theme-foo/4"}
// 	if id := card.ModelID(); id != expected {
// 		t.Errorf("Unexpected result: %d", id)
// 	}
// }

func TestCardNoteID(t *testing.T) {
	card := &Card{ID: "card-foo.bar.1"}
	expected := "note-bar"
	if id := card.NoteID(); id != expected {
		t.Errorf("Unexpected result: %s", id)
	}
}

func TestCardValidate(t *testing.T) {
	tests := []validationTest{
		{
			name: "empty card",
			v:    &Card{},
			err:  "id required",
		},
		{
			name: "invalid id",
			v:    &Card{ID: "chicken"},
			err:  "invalid ID type",
		},
		{
			name: "zero created time",
			v:    &Card{ID: "card-foo.bar.0"},
			err:  "created time required",
		},
		{
			name: "zero modified time",
			v:    &Card{ID: "card-foo.bar.0", Created: parseTime("2017-01-01T01:01:01Z")},
			err:  "modified time required",
		},
		{
			name: "missing model id",
			v:    &Card{ID: "card-foo.bar.0", Created: parseTime("2017-01-01T01:01:01Z"), Modified: parseTime("2017-01-01T01:01:01Z")},
			err:  "invalid theme ID type",
		},
		{
			name: "invalid model id",
			v: &Card{ID: "card-foo.bar.0", Created: parseTime("2017-01-01T01:01:01Z"), Modified: parseTime("2017-01-01T01:01:01Z"),
				ModelID: "chicken"},
			err: "invalid theme ID type",
		},
		{
			name: "valid",
			v: &Card{ID: "card-foo.bar.0", Created: parseTime("2017-01-01T01:01:01Z"), Modified: parseTime("2017-01-01T01:01:01Z"),
				ModelID: "theme-foo/2"},
		},
	}
	testValidation(t, tests)
}

func TestParseThemeID(t *testing.T) {
	type ptiTest struct {
		name  string
		card  *Card
		theme string
		model int
		err   string
	}
	tests := []ptiTest{
		{
			name: "no input",
			card: &Card{},
			err:  "invalid theme ID type",
		},
		{
			name: "invalid type",
			card: &Card{ModelID: "foo-bar"},
			err:  "invalid theme ID type",
		},
		{
			name: "too many parts",
			card: &Card{ModelID: "theme-bar/baz/qux"},
			err:  "invalid theme ID format",
		},
		{
			name: "invalid model id",
			card: &Card{ModelID: "theme-foo/bar"},
			err:  `invalid Model index: strconv.Atoi: parsing "bar": invalid syntax`,
		},
		{
			name:  "valid",
			card:  &Card{ModelID: "theme-foo/3"},
			theme: "foo",
			model: 3,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			theme, model, err := test.card.parseThemeID()
			checkErr(t, test.err, err)
			if err != nil {
				return
			}
			if theme != test.theme || model != test.model {
				t.Errorf("Unexpected result: %s %d", theme, model)
			}
		})
	}
}

func TestCardThemeID(t *testing.T) {
	card := &Card{ModelID: "theme-foo/2"}
	expected := "theme-foo"
	if id := card.ThemeID(); id != expected {
		t.Errorf("Unexpected result: %s", id)
	}
}

func TestCardThemeModelID(t *testing.T) {
	card := &Card{ModelID: "theme-foo/2"}
	expected := 2
	if id := card.ThemeModelID(); id != expected {
		t.Errorf("Unexpected result: %d", id)
	}
}
