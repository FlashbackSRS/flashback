package fb

import (
	"testing"

	"github.com/flimzy/diff"
)

func TestPkgValidate(t *testing.T) {
	tests := []validationTest{
		{
			name: "card without deck",
			err:  "card 'card-abcde.mViuXQThMLoh1G1Nlc4d_E8kR8o.0' found in package, but not in a deck",
			v: &Package{
				Cards: []*Card{
					{
						ID:       "card-abcde.mViuXQThMLoh1G1Nlc4d_E8kR8o.0",
						ModelID:  "theme-VGVzdCBUaGVtZQ/0",
						Created:  now(),
						Modified: now(),
					},
				},
			},
		},
		{
			name: "card missing from package",
			err:  "card 'card-abcde.mViuXQThMLoh1G1Nlc4d_E8kR8o.0' listed in deck, but not found in package",
			v: &Package{
				Decks: []*Deck{
					{
						ID:       "deck-AQID",
						Cards:    &CardCollection{map[string]struct{}{"card-abcde.mViuXQThMLoh1G1Nlc4d_E8kR8o.0": {}}},
						Created:  now(),
						Modified: now(),
					},
				},
			},
		},
		{
			name: "valid",
			v: &Package{
				Decks: []*Deck{
					{
						ID:       "deck-AQID",
						Cards:    &CardCollection{map[string]struct{}{"card-abcde.mViuXQThMLoh1G1Nlc4d_E8kR8o.0": {}}},
						Created:  now(),
						Modified: now(),
					},
				},
				Cards: []*Card{
					{
						ID:       "card-abcde.mViuXQThMLoh1G1Nlc4d_E8kR8o.0",
						ModelID:  "theme-VGVzdCBUaGVtZQ/0",
						Created:  now(),
						Modified: now(),
					},
				},
			},
		},
		{
			name: "note without matching model",
			err:  "note 'note-Zm9v' has no matching model (theme-Zm9v/3)",
			v: &Package{
				Notes: []*Note{
					{
						ID:          "note-Zm9v",
						ThemeID:     "theme-Zm9v",
						ModelID:     3,
						Created:     now(),
						Modified:    now(),
						Attachments: NewFileCollection(),
					},
				},
			},
		},
	}
	testValidation(t, tests)
}

func TestPkgMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		pkg      *Package
		expected string
		err      string
	}{
		{
			name: "empty package",
			pkg: &Package{
				Created:  now(),
				Modified: now(),
			},
			expected: `{"version":2, "created":"2017-01-01T00:00:00Z", "modified":"2017-01-01T00:00:00Z"}`,
		},
		{
			name: "invalid bundle",
			pkg:  &Package{Bundle: &Bundle{}},
			err:  "json: error calling MarshalJSON for type *fb.Bundle: id required",
		},
		{
			name: "invalid card",
			pkg:  &Package{Cards: []*Card{{}}},
			err:  "card '' validation': id required",
		},
		{
			name: "invalid note",
			pkg:  &Package{Notes: []*Note{{}}},
			err:  "note '' validation: id required",
		},
		{
			name: "invalid deck",
			pkg:  &Package{Decks: []*Deck{{}}},
			err:  "deck '' validation: id required",
		},
		{
			name: "invalid theme",
			pkg:  &Package{Themes: []*Theme{{}}},
			err:  "theme '' validation: id required",
		},
		{
			name: "invalid review",
			pkg:  &Package{Reviews: []*Review{{}}},
			err:  "json: error calling MarshalJSON for type *fb.Review: card id required",
		},
		{
			name: "full package",
			pkg: func() *Package {
				theme, _ := NewTheme("theme-abcd")
				model := &Model{
					Theme: theme,
					Type:  "foo",
					Files: theme.Attachments.NewView(),
				}
				theme.ModelSequence = 1
				theme.Models = []*Model{model}

				noteAtt := NewFileCollection()
				return &Package{
					Created:  now(),
					Modified: now(),
					Bundle:   &Bundle{ID: "bundle-mzxw6", Owner: "mjxwe", Created: now(), Modified: now()},
					Themes:   []*Theme{theme},
					Decks:    []*Deck{{ID: "deck-ZGVjaw", Created: now(), Modified: now(), Cards: &CardCollection{col: map[string]struct{}{"card-YmFy.bmlsCg.0": {}}}}},
					Notes:    []*Note{{ID: "note-Zm9v", ThemeID: "theme-abcd", Model: model, Created: now(), Modified: now(), Attachments: noteAtt}},
					Cards:    []*Card{{ID: "card-YmFy.bmlsCg.0", ModelID: "theme-abcd/0", Created: now(), Modified: now()}},
				}
			}(),
			expected: `{
				"version": 2,
				"created": "2017-01-01T00:00:00Z",
				"modified": "2017-01-01T00:00:00Z",
				"bundle": {"_id":"bundle-mzxw6", "type":"bundle", "owner":"mjxwe", "created":"2017-01-01T00:00:00Z", "modified":"2017-01-01T00:00:00Z"},
				"themes": [{"_id":"theme-abcd", "type":"theme", "created":"2017-01-01T00:00:00Z", "modified":"2017-01-01T00:00:00Z", "modelSequence":1, "files":[], "_attachments":{}, "models":[{"fields":null, "files":[], "modelType":"foo", "templates":null, "id":0}]}],
				"decks": [{"_id":"deck-ZGVjaw", "type":"deck", "created":"2017-01-01T00:00:00Z", "modified":"2017-01-01T00:00:00Z", "cards":["card-YmFy.bmlsCg.0"]}],
				"notes": [{"_id":"note-Zm9v", "type":"note", "created":"2017-01-01T00:00:00Z", "modified":"2017-01-01T00:00:00Z", "_attachments":{}, "fieldValues":null, "theme":"theme-abcd", "model":0}],
				"cards": [{"_id":"card-YmFy.bmlsCg.0", "type":"card", "created":"2017-01-01T00:00:00Z", "modified":"2017-01-01T00:00:00Z", "model": "theme-abcd/0"}]
			}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.pkg.MarshalJSON()
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

func TestPkgUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Package
		err      string
	}{
		{
			name:  "invalid json",
			input: "invalid json",
			err:   "invalid character 'i' looking for beginning of value",
		},
		{
			name:  "incompatible version",
			input: `{"version":0}`,
			err:   "package version 0 < 1",
		},
		{
			name:  "invalid bundle",
			input: `{"version":1, "bundle":{}}`,
			err:   "id required",
		},
		{
			name:  "invalid card",
			input: `{"version":1, "cards":[{"type":"card"}]}`,
			err:   "validation error: id required",
		},
		{
			name:  "invalid note",
			input: `{"version":1, "notes":[{"type":"note"}]}`,
			err:   "id required",
		},
		{
			name:  "invalid deck",
			input: `{"version":1, "decks":[{"type":"deck"}]}`,
			err:   "id required",
		},
		{
			name:  "invalid theme",
			input: `{"version":1, "themes":[{"type":"theme"}]}`,
			err:   "invalid theme: no attachments",
		},
		{
			name:  "invalid review",
			input: `{"version":1, "reviews":[{}]}`,
			err:   "card id required",
		},
		{
			name: "full package",
			input: `{
				"version": 1,
				"created": "2017-01-01T00:00:00Z",
				"modified": "2017-01-01T00:00:00Z",
				"bundle": {"_id":"bundle-mzxw6", "type":"bundle", "owner":"mjxwe", "created":"2017-01-01T00:00:00Z", "modified":"2017-01-01T00:00:00Z"},
				"themes": [{"_id":"theme-abcd", "type":"theme", "created":"2017-01-01T00:00:00Z", "modified":"2017-01-01T00:00:00Z", "modelSequence":1, "files":[], "_attachments":{}, "models":[{"fields":null, "files":[], "modelType":"foo", "templates":null, "id":0}]}],
				"decks": [{"_id":"deck-ZGVjaw", "type":"deck", "created":"2017-01-01T00:00:00Z", "modified":"2017-01-01T00:00:00Z", "cards":["card-YmFy.bmlsCg.0"]}],
				"notes": [{"_id":"note-Zm9v", "type":"note", "created":"2017-01-01T00:00:00Z", "modified":"2017-01-01T00:00:00Z", "_attachments":{}, "fieldValues":null, "theme":"theme-abcd", "model":0}],
				"cards": [{"_id":"card-YmFy.bmlsCg.0", "type":"card", "created":"2017-01-01T00:00:00Z", "modified":"2017-01-01T00:00:00Z", "model": "theme-abcd/0"}]
			}`,
			expected: func() *Package {
				theme, _ := NewTheme("theme-abcd")
				model := &Model{
					Theme: theme,
					Type:  "foo",
					Files: theme.Attachments.NewView(),
				}
				theme.ModelSequence = 1
				theme.Models = []*Model{model}

				noteAtt := NewFileCollection()
				return &Package{
					Created:  now(),
					Modified: now(),
					Bundle:   &Bundle{ID: "bundle-mzxw6", Owner: "mjxwe", Created: now(), Modified: now()},
					Themes:   []*Theme{theme},
					Decks:    []*Deck{{ID: "deck-ZGVjaw", Created: now(), Modified: now(), Cards: &CardCollection{col: map[string]struct{}{"card-YmFy.bmlsCg.0": {}}}}},
					Notes:    []*Note{{ID: "note-Zm9v", ThemeID: "theme-abcd", Model: model, Created: now(), Modified: now(), Attachments: noteAtt}},
					Cards:    []*Card{{ID: "card-YmFy.bmlsCg.0", ModelID: "theme-abcd/0", Created: now(), Modified: now()}},
				}
			}(),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := &Package{}
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
