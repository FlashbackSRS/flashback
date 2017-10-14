package model

import (
	"context"
	"errors"
	"sort"
	"testing"
	"time"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/testy"
)

func TestDeckList(t *testing.T) {
	tests := []struct {
		name     string
		repo     *Repo
		expected []*Deck
		err      string
	}{
		{
			name: "not logged in",
			repo: &Repo{},
			err:  "not logged in",
		},
		{
			name: "reduce query fail",
			repo: &Repo{
				user: "bob",
				local: &mockClient{
					db: &mockQuerier{err: errors.New("foo error")},
				},
			},
			err: "foo error",
		},
		{
			name: "fleshen error",
			repo: &Repo{
				user: "bob",
				local: &mockClient{
					db: &mockQuerier{
						kivikDB: &mockMultiGetter{rows: map[string]kivikRow{
							"deck-Brm5eFOpF0553VTksh7hlySt6M8": mockRow(`invalid json`),
						}},
						rows: []*mockRows{{
							rows:   []string{""},
							values: []string{"[234,6]"},
							keys:   []string{`["new","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`},
						}},
					},
				},
			},
			err: "invalid character 'i' looking for beginning of value",
		},
		{
			name: "success",
			repo: &Repo{
				user: "bob",
				local: &mockClient{
					db: &mockQuerier{
						kivikDB: &mockMultiGetter{rows: map[string]kivikRow{
							"deck-Brm5eFOpF0553VTksh7hlySt6M8": mockRow(`{"name":"Test Deck"}`),
							"deck-foo":                         mockRow(`{"name":"Foo"}`),
							"deck-bar":                         mockRow(`{"name":"Bar"}`),
						}},
						options: []kivik.Options{
							{"group_level": 2},
							{"startkey": []interface{}{"old", "deck-Brm5eFOpF0553VTksh7hlySt6M8"}, "reduce": false},
							{"startkey": []interface{}{"old", "deck-foo"}, "reduce": false},
						},
						rows: []*mockRows{
							{
								rows: []string{"", "", "", "", "", "", ""},
								values: []string{
									"[234,6]", "[1811,56]", "[52,9]",
									"[100,0]", "[100,20]", "[5,1]",
									"[50,0]",
								},
								keys: []string{
									`["new","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`,
									`["old","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`,
									`["suspended","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`,
									`["new","deck-foo"]`,
									`["old","deck-foo"]`,
									`["suspended","deck-foo"]`,
									`["new","deck-bar"]`,
								},
							},
							{rows: []string{"", "", "", ""}, values: []string{"{}", "{}", "{}", "{}"}},
							{rows: []string{"", ""}, values: []string{"{}", "{}"}},
						},
					},
				},
			},
			expected: []*Deck{
				{
					Name:           "All",
					TotalCards:     2352,
					DueCards:       6,
					LearningCards:  76,
					MatureCards:    1835,
					NewCards:       384,
					SuspendedCards: 57,
				},
				{
					Name:       "Bar",
					ID:         "deck-bar",
					TotalCards: 50,
					NewCards:   50,
				},
				{
					Name:           "Foo",
					ID:             "deck-foo",
					TotalCards:     205,
					DueCards:       2,
					LearningCards:  20,
					MatureCards:    80,
					NewCards:       100,
					SuspendedCards: 5,
				},
				{
					Name:           "Test Deck",
					ID:             "deck-Brm5eFOpF0553VTksh7hlySt6M8",
					TotalCards:     2097,
					DueCards:       4,
					LearningCards:  56,
					MatureCards:    1755,
					NewCards:       234,
					SuspendedCards: 52,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.repo.DeckList(context.Background())
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != test.err {
				t.Errorf("Unexpected error: %s", errMsg)
			}
			if err != nil {
				return
			}
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestFleshenDecks(t *testing.T) {
	tests := []struct {
		name     string
		db       kivikDB
		input    []*Deck
		err      string
		expected []*Deck
	}{
		{
			name:  "name get error",
			db:    &mockGetter{err: errors.New("get error")},
			input: []*Deck{{ID: "deck-foo"}},
			err:   "get error",
		},
		{
			name: "dueCount error",
			db: &mockQuerier{
				kivikDB: &mockGetter{row: mockRow(`{}`)},
				err:     errors.New("query error"),
			},
			input: []*Deck{{ID: "deck-foo", MatureCards: 10}},
			err:   "query error",
		},
		{
			name: "success",
			db: &mockQuerier{
				kivikDB: &mockMultiGetter{
					rows: map[string]kivikRow{
						"deck-foo": mockRow(`{"name":"Foo"}`),
						"deck-bar": mockRow(`{"name":"Bar"}`),
						"deck-baz": mockRow(`{"name":"Baz"}`),
					},
				},
				options: []kivik.Options{
					{"startkey": []interface{}{"old", "deck-bar"}},
					{"startkey": []interface{}{"old", "deck-baz"}},
				},
				rows: []*mockRows{
					{
						rows:   []string{""},
						values: []string{`{}`},
					},
					{
						rows:   []string{"", "", ""},
						values: []string{`{}`, `{}`, `{}`},
					},
				},
			},
			input: []*Deck{
				{ID: "deck-foo"},
				{ID: "deck-bar", LearningCards: 10},
				{ID: "deck-baz", MatureCards: 10},
			},
			expected: []*Deck{
				{ID: "deck-foo", Name: "Foo"},
				{ID: "deck-bar", Name: "Bar", LearningCards: 10, DueCards: 1},
				{ID: "deck-baz", Name: "Baz", MatureCards: 10, DueCards: 3},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := fleshenDecks(context.Background(), test.db, test.input)
			testy.Error(t, test.err, err)
			if d := diff.Interface(test.expected, test.input); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestDeckReducedStats(t *testing.T) {
	tests := []struct {
		name     string
		db       querier
		expected []*Deck
		err      string
	}{
		{
			name: "query error",
			db:   &mockQuerier{err: errors.New("some error")},
			err:  "some error",
		},
		{
			name:     "no cards",
			db:       &mockQuerier{rows: []*mockRows{{}}},
			expected: []*Deck{},
		},
		{
			name: "one deck",
			db: &mockQuerier{
				rows: []*mockRows{{
					rows:   []string{"", "", ""},
					values: []string{"[234,6]", "[1811,56]", "[52,9]"},
					keys: []string{
						`["new","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`,
						`["old","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`,
						`["suspended","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`,
					},
				}},
			},
			expected: []*Deck{
				{
					ID:             "deck-Brm5eFOpF0553VTksh7hlySt6M8",
					TotalCards:     2097,
					DueCards:       0,
					LearningCards:  56,
					MatureCards:    1755,
					NewCards:       234,
					SuspendedCards: 52,
				},
			},
		},
		{
			name: "multiple decks",
			db: &mockQuerier{
				rows: []*mockRows{{
					rows: []string{"", "", "", "", "", "", ""},
					values: []string{
						"[234,6]", "[1811,56]", "[52,9]",
						"[100,0]", "[100,20]", "[5,1]",
						"[50,0]",
					},
					keys: []string{
						`["new","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`,
						`["old","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`,
						`["suspended","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`,
						`["new","deck-foo"]`,
						`["old","deck-foo"]`,
						`["suspended","deck-foo"]`,
						`["new","deck-bar"]`,
					},
				}},
			},
			expected: []*Deck{
				{
					ID:             "deck-Brm5eFOpF0553VTksh7hlySt6M8",
					TotalCards:     2097,
					DueCards:       0,
					LearningCards:  56,
					MatureCards:    1755,
					NewCards:       234,
					SuspendedCards: 52,
				},
				{
					ID:         "deck-bar",
					TotalCards: 50,
					NewCards:   50,
				},
				{
					ID:             "deck-foo",
					TotalCards:     205,
					DueCards:       0,
					LearningCards:  20,
					MatureCards:    80,
					NewCards:       100,
					SuspendedCards: 5,
				},
			},
		},
		{
			name: "invalid json value",
			db: &mockQuerier{
				rows: []*mockRows{{
					rows:   []string{""},
					values: []string{"invalid json"},
					keys:   []string{`["new","deck-Brm5eFOpF0553VTksh7hlySt6M8"]`},
				}},
			},
			err: "invalid character 'i' looking for beginning of value",
		},
		{
			name: "invalid json key",
			db: &mockQuerier{
				rows: []*mockRows{{
					rows:   []string{""},
					values: []string{"[10,10]"},
					keys:   []string{`invalid json`},
				}},
			},
			err: "invalid character 'i' looking for beginning of value",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := deckReducedStats(context.Background(), test.db)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if test.err != errMsg {
				t.Errorf("Unexpected error: %s", errMsg)
			}
			if err != nil {
				return
			}
			sort.Slice(result, func(i, j int) bool {
				return result[i].ID < result[j].ID
			})
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestDueCount(t *testing.T) {
	tests := []struct {
		name     string
		db       querier
		deckID   string
		ts       time.Time
		expected int
		err      string
	}{
		{
			name: "query error",
			db:   &mockQuerier{err: errors.New("unf error")},
			err:  "unf error",
		},
		{
			name:     "no results",
			db:       &mockQuerier{rows: []*mockRows{{}}},
			deckID:   "deck-foo",
			ts:       now(),
			expected: 0,
		},
		{
			name: "some results",
			db: &mockQuerier{
				options: []kivik.Options{
					{
						"startkey":     []interface{}{"old", "deck-foo"},
						"endkey":       []interface{}{"old", "deck-foo", "2017-01-01T12:00:00Z"},
						"reduce":       false,
						"include_docs": false,
					},
				},
				rows: []*mockRows{
					{rows: []string{"", "", ""}, values: []string{"{}", "{}", "{}"}},
				},
			},
			deckID:   "deck-foo",
			ts:       parseTime(t, "2017-01-01T12:00:00Z"),
			expected: 3,
		},
		{
			name: "value error",
			db: &mockQuerier{
				options: []kivik.Options{
					{
						"startkey":     []interface{}{"old", "deck-foo"},
						"endkey":       []interface{}{"old", "deck-foo", "2017-01-01T12:00:00Z"},
						"reduce":       false,
						"include_docs": false,
					},
				},
				rows: []*mockRows{
					{rows: []string{""}, values: []string{`invalid json`}},
				},
			},
			deckID: "deck-foo",
			ts:     parseTime(t, "2017-01-01T12:00:00Z"),
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name: "Omit buried",
			db: &mockQuerier{
				options: []kivik.Options{
					{
						"startkey":     []interface{}{"old", "deck-foo"},
						"endkey":       []interface{}{"old", "deck-foo", "2017-01-01T12:00:00Z"},
						"reduce":       false,
						"include_docs": false,
					},
				},
				rows: []*mockRows{
					{rows: []string{"", "", ""}, values: []string{`{"buriedUntil":"2019-01-01"}`, "{}", "{}"}},
				},
			},
			deckID:   "deck-foo",
			ts:       parseTime(t, "2017-01-01T12:00:00Z"),
			expected: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := dueCount(context.Background(), test.db, test.deckID, test.ts)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != test.err {
				t.Errorf("Unexpected error: %s", errMsg)
			}
			if err != nil {
				return
			}
			if test.expected != result {
				t.Errorf("Unexpected result: %d", result)
			}
		})
	}
}

func TestDeckName(t *testing.T) {
	tests := []struct {
		name     string
		db       getter
		deckID   string
		expected string
		err      string
	}{
		{
			name: "get error",
			db:   &mockQueryGetter{err: errors.New("error getting")},
			err:  "error getting",
		},
		{
			name: "bad json",
			db:   &mockQueryGetter{row: mockRow("invalid json")},
			err:  "invalid character 'i' looking for beginning of value",
		},
		{
			name:     "success",
			db:       &mockQueryGetter{row: mockRow(`{"name":"foo deck","unused":"field"}`)},
			expected: "foo deck",
		},
		{
			name:     "orphan deck",
			deckID:   orphanedCardDeckID,
			expected: orphanedCardDeckName,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := deckName(context.Background(), test.db, test.deckID)
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != test.err {
				t.Errorf("Unexpected error: %s", errMsg)
			}
			if err != nil {
				return
			}
			if result != test.expected {
				t.Errorf("Unexpected result: %s", result)
			}
		})
	}
}
