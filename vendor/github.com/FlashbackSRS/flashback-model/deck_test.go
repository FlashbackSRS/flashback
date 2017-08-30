package fb

import (
	"testing"

	"github.com/flimzy/diff"
)

func TestCCMarshalJSON(t *testing.T) {
	type Test struct {
		name     string
		cc       *CardCollection
		expected string
		err      string
	}
	tests := []Test{
		{
			name:     "empty",
			cc:       &CardCollection{},
			expected: "[]",
		},
		{
			name: "some cards",
			cc: &CardCollection{
				col: map[string]struct{}{
					"card-Zm9v.bmlsCg.0": {},
					"card-YmFy.bmlsCg.0": {},
				},
			},
			expected: `["card-YmFy.bmlsCg.0","card-Zm9v.bmlsCg.0"]`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.cc.MarshalJSON()
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

func TestNewCardCollection(t *testing.T) {
	cc := NewCardCollection()
	expected := &CardCollection{
		col: map[string]struct{}{},
	}
	if d := diff.Interface(expected, cc); d != nil {
		t.Error(d)
	}
}

func TestCCUnmarshalJSON(t *testing.T) {
	type Test struct {
		name     string
		input    string
		expected *CardCollection
		err      string
	}
	tests := []Test{
		{
			name:  "invalid json",
			input: "invalid json",
			err:   "invalid character 'i' looking for beginning of value",
		},
		{
			name:  "valid",
			input: `["card-Zm9v.bmlsCg.0","card-YmFy.bmlsCg.0"]`,
			expected: &CardCollection{col: map[string]struct{}{
				"card-Zm9v.bmlsCg.0": {},
				"card-YmFy.bmlsCg.0": {},
			}},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := &CardCollection{}
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

func TestCCAll(t *testing.T) {
	cc := &CardCollection{col: map[string]struct{}{
		"card-Zm9v.bmlsCg.0": {},
		"card-YmFy.bmlsCg.0": {},
	}}
	expected := []string{"card-YmFy.bmlsCg.0", "card-Zm9v.bmlsCg.0"}
	result := cc.All()
	if d := diff.Interface(expected, result); d != nil {
		t.Error(d)
	}
}

func TestNewDeck(t *testing.T) {
	type Test struct {
		name     string
		id       string
		expected *Deck
		err      string
	}
	tests := []Test{
		{
			name: "no id",
			err:  "id required",
		},
		{
			name: "valid",
			id:   "deck-Zm9vIGlkCg",
			expected: &Deck{
				ID:       "deck-Zm9vIGlkCg",
				Created:  now(),
				Modified: now(),
				Cards:    &CardCollection{col: map[string]struct{}{}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := NewDeck(test.id)
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

func TestDeckMarshalJSON(t *testing.T) {
	type Test struct {
		name     string
		deck     *Deck
		expected string
		err      string
	}
	tests := []Test{
		{
			name: "invalid deck",
			deck: &Deck{},
			err:  "id required",
		},
		{
			name: "full fields",
			deck: &Deck{
				ID:          "deck-ZGVjaw",
				Created:     now(),
				Modified:    now(),
				Imported:    now(),
				Name:        "test name",
				Description: "test description",
				Cards: &CardCollection{col: map[string]struct{}{
					"card-Zm9v.bmlsCg.0": {}, "card-YmFy.bmlsCg.0": {},
				}},
			},
			expected: `{
				"_id":         "deck-ZGVjaw",
				"type":        "deck",
				"name":        "test name",
				"description": "test description",
				"created":     "2017-01-01T00:00:00Z",
				"modified":    "2017-01-01T00:00:00Z",
				"imported":    "2017-01-01T00:00:00Z",
				"cards":       ["card-YmFy.bmlsCg.0","card-Zm9v.bmlsCg.0"]
			}`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.deck.MarshalJSON()
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

func TestDeckAddCard(t *testing.T) {
	deck, err := NewDeck("deck-Zm9v")
	if err != nil {
		t.Fatal(err)
	}
	deck.AddCard("card-jack")
	deck.AddCard("card-jill")
	expected := &Deck{
		ID:       "deck-Zm9v",
		Created:  now(),
		Modified: now(),
		Cards: &CardCollection{col: map[string]struct{}{
			"card-jack": {},
			"card-jill": {},
		}},
	}
	if d := diff.Interface(expected, deck); d != nil {
		t.Error(d)
	}
}

func TestDeckUnmarshalJSON(t *testing.T) {
	type Test struct {
		name     string
		input    string
		expected *Deck
		err      string
	}
	tests := []Test{
		{
			name:  "invalid json",
			input: "invalid json",
			err:   "invalid character 'i' looking for beginning of value",
		},
		{
			name:  "invalid deck",
			input: `{}`,
			err:   "id required",
		},
		{
			name: "all fields",
			input: `{
                "_id":         "deck-ZGVjaw",
                "name":        "test name",
                "description": "test description",
                "created":     "2017-01-01T00:00:00Z",
                "modified":    "2017-01-01T00:00:00Z",
                "imported":    "2017-01-01T00:00:00Z",
                "cards":       ["card-YmFy.bmlsCg.0","card-Zm9v.bmlsCg.0"]
            }`,
			expected: &Deck{
				ID:          "deck-ZGVjaw",
				Created:     now(),
				Modified:    now(),
				Imported:    now(),
				Name:        "test name",
				Description: "test description",
				Cards:       &CardCollection{col: map[string]struct{}{"card-YmFy.bmlsCg.0": {}, "card-Zm9v.bmlsCg.0": {}}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := &Deck{}
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

func TestDeckSetRev(t *testing.T) {
	deck := &Deck{}
	rev := "1-xxx"
	deck.SetRev(rev)
	if deck.Rev != rev {
		t.Errorf("Unexpected result: %s", deck.Rev)
	}
}

func TestDeckDocID(t *testing.T) {
	deck := &Deck{ID: "deck-Zm9v"}
	expected := "deck-Zm9v"
	if id := deck.DocID(); id != expected {
		t.Errorf("Unexpected result: %s", id)
	}
}

func TestDeckImportedTime(t *testing.T) {
	t.Run("Set", func(t *testing.T) {
		ts := now()
		deck := &Deck{Imported: ts}
		if it := deck.ImportedTime(); it != ts {
			t.Errorf("Unexpected result: %s", it)
		}
	})
	t.Run("Unset", func(t *testing.T) {
		deck := &Deck{}
		if it := deck.ImportedTime(); !it.IsZero() {
			t.Errorf("unexpected result: %v", it)
		}
	})
}

func TestDeckModifiedTime(t *testing.T) {
	deck := &Deck{}
	ts := now()
	deck.Modified = ts
	if mt := deck.ModifiedTime(); mt != ts {
		t.Errorf("Unexpected result")
	}
}

func TestDeckMergeImport(t *testing.T) {
	type Test struct {
		name         string
		new          *Deck
		existing     *Deck
		expected     bool
		expectedDeck *Deck
		err          string
	}
	tests := []Test{
		{
			name:     "different ids",
			new:      &Deck{ID: "deck-YWJjZAo"},
			existing: &Deck{ID: "deck-YWJjZQo"},
			err:      "IDs don't match",
		},
		{
			name:     "created timestamps don't match",
			new:      &Deck{ID: "deck-YWJjZAo", Created: parseTime("2017-01-01T01:01:01Z"), Imported: parseTime("2017-01-15T00:00:00Z")},
			existing: &Deck{ID: "deck-YWJjZAo", Created: parseTime("2017-02-01T01:01:01Z"), Imported: parseTime("2017-01-20T00:00:00Z")},
			err:      "Created timestamps don't match",
		},
		{
			name:     "new not an import",
			new:      &Deck{ID: "deck-YWJjZAo", Created: parseTime("2017-01-01T01:01:01Z")},
			existing: &Deck{ID: "deck-YWJjZAo", Created: parseTime("2017-01-01T01:01:01Z"), Imported: parseTime("2017-01-15T00:00:00Z")},
			err:      "not an import",
		},
		{
			name:     "existing not an import",
			new:      &Deck{ID: "deck-YWJjZAo", Created: parseTime("2017-01-01T01:01:01Z"), Imported: parseTime("2017-01-15T00:00:00Z")},
			existing: &Deck{ID: "deck-YWJjZAo", Created: parseTime("2017-01-01T01:01:01Z")},
			err:      "not an import",
		},
		{
			name: "new is newer",
			new: &Deck{
				ID:          "deck-YWJjZAo",
				Name:        "foo",
				Description: "FOO",
				Created:     parseTime("2017-01-01T01:01:01Z"),
				Modified:    parseTime("2017-02-01T01:01:01Z"),
				Imported:    parseTime("2017-01-15T00:00:00Z"),
				Cards:       &CardCollection{col: map[string]struct{}{"card-Zm9v.bmlsCg.0": {}}},
			},
			existing: &Deck{
				ID:          "deck-YWJjZAo",
				Name:        "bar",
				Description: "BAR",
				Created:     parseTime("2017-01-01T01:01:01Z"),
				Modified:    parseTime("2017-01-01T01:01:01Z"),
				Imported:    parseTime("2017-01-20T00:00:00Z"),
				Cards:       &CardCollection{col: map[string]struct{}{"card-YmFy.bmlsCg.0": {}}},
			},
			expected: true,
			expectedDeck: &Deck{
				ID:          "deck-YWJjZAo",
				Name:        "foo",
				Description: "FOO",
				Created:     parseTime("2017-01-01T01:01:01Z"),
				Modified:    parseTime("2017-02-01T01:01:01Z"),
				Imported:    parseTime("2017-01-15T00:00:00Z"),
				Cards:       &CardCollection{col: map[string]struct{}{"card-Zm9v.bmlsCg.0": {}}},
			},
		},
		{
			name: "existing is newer",
			new: &Deck{
				ID:          "deck-YWJjZAo",
				Name:        "foo",
				Description: "FOO",
				Created:     parseTime("2017-01-01T01:01:01Z"),
				Modified:    parseTime("2017-01-01T01:01:01Z"),
				Imported:    parseTime("2017-01-15T00:00:00Z"),
				Cards:       &CardCollection{col: map[string]struct{}{"card-Zm9v.bmlsCg.0": {}}},
			},
			existing: &Deck{
				ID:          "deck-YWJjZAo",
				Name:        "bar",
				Description: "BAR",
				Created:     parseTime("2017-01-01T01:01:01Z"),
				Modified:    parseTime("2017-02-01T01:01:01Z"),
				Imported:    parseTime("2017-01-20T00:00:00Z"),
				Cards:       &CardCollection{col: map[string]struct{}{"card-YmFy.bmlsCg.0": {}}},
			},
			expected: false,
			expectedDeck: &Deck{
				ID:          "deck-YWJjZAo",
				Name:        "bar",
				Description: "BAR",
				Created:     parseTime("2017-01-01T01:01:01Z"),
				Modified:    parseTime("2017-02-01T01:01:01Z"),
				Imported:    parseTime("2017-01-20T00:00:00Z"),
				Cards:       &CardCollection{col: map[string]struct{}{"card-YmFy.bmlsCg.0": {}}},
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
			if d := diff.Interface(test.expectedDeck, test.new); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestCCValidate(t *testing.T) {
	tests := []validationTest{
		{
			name: "no cards",
			v:    &CardCollection{},
		},
		{
			name: "invalid card ID",
			v:    &CardCollection{col: map[string]struct{}{"foo": {}}},
			err:  "'foo': invalid ID type",
		},
	}
	testValidation(t, tests)
}

func TestDeckValidate(t *testing.T) {
	tests := []validationTest{
		{
			name: "no id",
			v:    &Deck{},
			err:  "id required",
		},
		{
			name: "invalid doctype",
			v:    &Deck{ID: "chicken-abcd"},
			err:  "unsupported DocID type 'chicken'",
		},
		{
			name: "invalid doctype",
			v:    &Deck{ID: "theme-abcd"},
			err:  "incorrect doc type",
		},
		{
			name: "no created time",
			v:    &Deck{ID: "deck-YWJjZAo"},
			err:  "created time required",
		},
		{
			name: "no modified time",
			v:    &Deck{ID: "deck-YWJjZAo", Created: now()},
			err:  "modified time required",
		},
		{
			name: "nil collection",
			v:    &Deck{ID: "deck-YWJjZAo", Created: now(), Modified: now()},
			err:  "collection is nil",
		},
		{
			name: "invalid card",
			v:    &Deck{ID: "deck-YWJjZAo", Created: now(), Modified: now(), Cards: &CardCollection{col: map[string]struct{}{"foo": {}}}},
			err:  "'foo': invalid ID type",
		},
		{
			name: "valid",
			v:    &Deck{ID: "deck-YWJjZAo", Created: now(), Modified: now(), Cards: &CardCollection{col: map[string]struct{}{"card-abcd.abcd.0": {}}}},
		},
	}
	testValidation(t, tests)
}
